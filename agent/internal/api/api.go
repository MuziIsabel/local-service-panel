// Package api provides the HTTP API handlers for the Agent.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/user/local-service-panel/agent/internal/domain"
	"github.com/user/local-service-panel/agent/internal/version"
	"github.com/user/local-service-panel/agent/internal/windowsservice"
	"github.com/user/local-service-panel/agent/internal/customapp"
	"github.com/user/local-service-panel/agent/internal/autostart"
	"github.com/user/local-service-panel/agent/internal/events"
	"github.com/user/local-service-panel/agent/internal/settings"
)

// APIResponse is the standard success response wrapper.
type APIResponse struct {
	Data interface{} `json:"data"`
}

// APIError is the standard error response wrapper.
type APIError struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains error details.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// HealthzResponse is the response for GET /api/healthz.
type HealthzResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ServiceActionResponse is the response for service operations (start/stop/restart).
type ServiceActionResponse struct {
	ServiceName string `json:"serviceName"`
	Status      string `json:"status"`
}

// SetStartTypeRequest is the request body for PATCH .../start-type.
type SetStartTypeRequest struct {
	StartType string `json:"startType"`
}

// SetStartTypeResponse is the response for PATCH .../start-type.
type SetStartTypeResponse struct {
	ServiceName string `json:"serviceName"`
	StartType   string `json:"startType"`
}

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	svcProvider   windowsservice.Provider
	customAppSvc  *customapp.Service
	eventSvc      *events.Service
	settingsStore *settings.Store
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(provider windowsservice.Provider, appSvc *customapp.Service, evtSvc *events.Service, settingsStore *settings.Store) *Handler {
	return &Handler{svcProvider: provider, customAppSvc: appSvc, eventSvc: evtSvc, settingsStore: settingsStore}
}

// Router creates and returns the main HTTP handler with all routes registered.
// Uses Go 1.22+ ServeMux patterns for method and path variable routing.
func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", h.handleHealthz)
	mux.HandleFunc("GET /api/windows/services", h.handleListServices)
	mux.HandleFunc("GET /api/windows/services/{serviceName}", h.handleGetService)
	mux.HandleFunc("POST /api/windows/services/{serviceName}/start", h.handleStartService)
	mux.HandleFunc("POST /api/windows/services/{serviceName}/stop", h.handleStopService)
	mux.HandleFunc("POST /api/windows/services/{serviceName}/restart", h.handleRestartService)
	mux.HandleFunc("PATCH /api/windows/services/{serviceName}/start-type", h.handleSetStartType)
	mux.HandleFunc("GET /api/custom-apps", h.handleListCustomApps)
	mux.HandleFunc("GET /api/custom-apps/{id}", h.handleGetCustomApp)
	mux.HandleFunc("POST /api/custom-apps", h.handleCreateCustomApp)
	mux.HandleFunc("PATCH /api/custom-apps/{id}", h.handleUpdateCustomApp)
	mux.HandleFunc("DELETE /api/custom-apps/{id}", h.handleDeleteCustomApp)
	mux.HandleFunc("POST /api/custom-apps/{id}/start", h.handleStartCustomApp)
	mux.HandleFunc("POST /api/custom-apps/{id}/stop", h.handleStopCustomApp)
	mux.HandleFunc("GET /api/custom-apps/{id}/logs", h.handleGetCustomAppLogs)
	mux.HandleFunc("POST /api/custom-apps/{id}/autostart", h.handleSetCustomAppAutoStart)
	mux.HandleFunc("GET /api/events", h.handleListEvents)
	mux.HandleFunc("GET /api/settings", h.handleGetSettings)
	mux.HandleFunc("PATCH /api/settings", h.handlePatchSettings)
	mux.HandleFunc("GET /api/targets", h.handleListTargets)
	mux.HandleFunc("GET /api/targets/{id}", h.handleGetTarget)
	mux.HandleFunc("POST /api/targets/{id}/start", h.handleStartTarget)
	mux.HandleFunc("POST /api/targets/{id}/stop", h.handleStopTarget)
	mux.HandleFunc("POST /api/targets/{id}/restart", h.handleRestartTarget)
	mux.HandleFunc("POST /api/targets/{id}/autostart", h.handleSetTargetAutoStart)
	mux.HandleFunc("GET /api/targets/{id}/logs", h.handleGetTargetLogs)
	mux.HandleFunc("/api/", h.handleNotFound)
	return mux
}

// handleNotFound returns 404 for unimplemented API routes.
func (h *Handler) handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "NOT_FOUND", "The requested resource was not found")
}

// handleHealthz handles GET /api/healthz.
func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse{
		Data: HealthzResponse{
			Status:  "ok",
			Version: version.Version,
		},
	})
}

// handleListServices handles GET /api/windows/services.
func (h *Handler) handleListServices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := windowsservice.Filter{
		Keyword:          q.Get("keyword"),
		Status:           domain.ServiceStatus(q.Get("status")),
		StartType:        domain.StartType(q.Get("startType")),
		IncludeProtected: q.Get("includeProtected") != "false",
	}

	services, err := h.svcProvider.List(r.Context(), filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	dtos := windowsservice.ToDTOList(services, windowsservice.IsProtected)
	writeJSON(w, http.StatusOK, APIResponse{Data: dtos})
}

// handleGetService handles GET /api/windows/services/{serviceName}.
func (h *Handler) handleGetService(w http.ResponseWriter, r *http.Request) {
	serviceName := r.PathValue("serviceName")

	svc, err := h.svcProvider.Get(r.Context(), serviceName)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Data: windowsservice.ToDTO(*svc, windowsservice.IsProtected(svc.ServiceName)),
	})
}

// handleStartService handles POST /api/windows/services/{serviceName}/start.
func (h *Handler) handleStartService(w http.ResponseWriter, r *http.Request) {
	serviceName := r.PathValue("serviceName")

	if err := h.svcProvider.Start(r.Context(), serviceName); err != nil {
		h.eventSvc.Record(serviceName, "windows_service", "WINDOWS_SERVICE_START_FAILED", "failed", err.Error(), "")
		writeServiceError(w, err)
		return
	}

	// Query the service to get its current status.
	svc, err := h.svcProvider.Get(r.Context(), serviceName)
	if err != nil {
		h.eventSvc.Record(serviceName, "windows_service", "WINDOWS_SERVICE_START_FAILED", "failed", err.Error(), "")
		writeServiceError(w, err)
		return
	}

	h.eventSvc.Record(serviceName, "windows_service", "WINDOWS_SERVICE_STARTED", "success",
		fmt.Sprintf("Started service %q", svc.DisplayName), "")

	writeJSON(w, http.StatusOK, APIResponse{
		Data: ServiceActionResponse{
			ServiceName: string(svc.ServiceName),
			Status:      string(svc.Status),
		},
	})
}

// handleStopService handles POST /api/windows/services/{serviceName}/stop.
func (h *Handler) handleStopService(w http.ResponseWriter, r *http.Request) {
	serviceName := r.PathValue("serviceName")

	if err := h.svcProvider.Stop(r.Context(), serviceName); err != nil {
		h.eventSvc.Record(serviceName, "windows_service", "WINDOWS_SERVICE_STOP_FAILED", "failed", err.Error(), "")
		writeServiceError(w, err)
		return
	}

	svc, err := h.svcProvider.Get(r.Context(), serviceName)
	if err != nil {
		h.eventSvc.Record(serviceName, "windows_service", "WINDOWS_SERVICE_STOP_FAILED", "failed", err.Error(), "")
		writeServiceError(w, err)
		return
	}

	h.eventSvc.Record(serviceName, "windows_service", "WINDOWS_SERVICE_STOPPED", "success",
		fmt.Sprintf("Stopped service %q", svc.DisplayName), "")

	writeJSON(w, http.StatusOK, APIResponse{
		Data: ServiceActionResponse{
			ServiceName: string(svc.ServiceName),
			Status:      string(svc.Status),
		},
	})
}

// handleRestartService handles POST /api/windows/services/{serviceName}/restart.
func (h *Handler) handleRestartService(w http.ResponseWriter, r *http.Request) {
	serviceName := r.PathValue("serviceName")

	if err := h.svcProvider.Restart(r.Context(), serviceName); err != nil {
		h.eventSvc.Record(serviceName, "windows_service", "WINDOWS_SERVICE_RESTART_FAILED", "failed", err.Error(), "")
		writeServiceError(w, err)
		return
	}

	svc, err := h.svcProvider.Get(r.Context(), serviceName)
	if err != nil {
		h.eventSvc.Record(serviceName, "windows_service", "WINDOWS_SERVICE_RESTART_FAILED", "failed", err.Error(), "")
		writeServiceError(w, err)
		return
	}

	h.eventSvc.Record(serviceName, "windows_service", "WINDOWS_SERVICE_RESTARTED", "success",
		fmt.Sprintf("Restarted service %q", svc.DisplayName), "")

	writeJSON(w, http.StatusOK, APIResponse{
		Data: ServiceActionResponse{
			ServiceName: string(svc.ServiceName),
			Status:      string(svc.Status),
		},
	})
}

// handleSetStartType handles PATCH /api/windows/services/{serviceName}/start-type.
func (h *Handler) handleSetStartType(w http.ResponseWriter, r *http.Request) {
	serviceName := r.PathValue("serviceName")

	var req SetStartTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	startType, err := windowsservice.ParseStartType(req.StartType)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	if err := h.svcProvider.SetStartType(r.Context(), serviceName, startType); err != nil {
		h.eventSvc.Record(serviceName, "windows_service", "WINDOWS_SERVICE_START_TYPE_CHANGE_FAILED", "failed", err.Error(), "")
		writeServiceError(w, err)
		return
	}

	h.eventSvc.Record(serviceName, "windows_service", "WINDOWS_SERVICE_START_TYPE_CHANGED", "success",
		fmt.Sprintf("Changed start type of service %q to %s", serviceName, req.StartType), "")

	writeJSON(w, http.StatusOK, APIResponse{
		Data: SetStartTypeResponse{
			ServiceName: serviceName,
			StartType:   req.StartType,
		},
	})
}

// handleListCustomApps handles GET /api/custom-apps.
func (h *Handler) handleListCustomApps(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")

	apps, err := h.customAppSvc.List(keyword)
	if err != nil {
		writeCustomAppError(w, err)
		return
	}

	dtos := customapp.ToDTOList(apps)
	if dtos == nil {
		dtos = []customapp.DTO{}
	}
	writeJSON(w, http.StatusOK, APIResponse{Data: dtos})
}

// handleGetCustomApp handles GET /api/custom-apps/{id}.
func (h *Handler) handleGetCustomApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	app, err := h.customAppSvc.GetByID(id)
	if err != nil {
		writeCustomAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Data: customapp.ToDTO(app)})
}

// handleCreateCustomApp handles POST /api/custom-apps.
func (h *Handler) handleCreateCustomApp(w http.ResponseWriter, r *http.Request) {
	var req customapp.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	app, err := h.customAppSvc.Create(&req)
	if err != nil {
		h.eventSvc.Record("", "custom_app", "CUSTOM_APP_CREATE_FAILED", "failed", err.Error(), "")
		writeCustomAppError(w, err)
		return
	}

	h.eventSvc.Record(app.ID, "custom_app", "CUSTOM_APP_CREATED", "success",
		fmt.Sprintf("Created custom app %q", app.Name), "")

	writeJSON(w, http.StatusCreated, APIResponse{Data: customapp.ToDTO(app)})
}

// handleStartCustomApp handles POST /api/custom-apps/{id}/start.
func (h *Handler) handleStartCustomApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	app, err := h.customAppSvc.Start(id)
	if err != nil {
		h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_START_FAILED", "failed", err.Error(), "")
		writeCustomAppError(w, err)
		return
	}

	h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_STARTED", "success",
		fmt.Sprintf("Started custom app %q (PID %d)", app.Name, app.PID), "")

	writeJSON(w, http.StatusOK, APIResponse{Data: customapp.ToDTO(app)})
}

// handleStopCustomApp handles POST /api/custom-apps/{id}/stop.
func (h *Handler) handleStopCustomApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	app, err := h.customAppSvc.Stop(id)
	if err != nil {
		h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_STOP_FAILED", "failed", err.Error(), "")
		writeCustomAppError(w, err)
		return
	}

	h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_STOPPED", "success",
		fmt.Sprintf("Stopped custom app %q", app.Name), "")

	writeJSON(w, http.StatusOK, APIResponse{Data: customapp.ToDTO(app)})
}

// handleGetCustomAppLogs handles GET /api/custom-apps/{id}/logs.
func (h *Handler) handleGetCustomAppLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Parse lines query parameter, default 200
	lines := 200
	if l := r.URL.Query().Get("lines"); l != "" {
		if parsed, err := parseInt(l); err == nil && parsed > 0 {
			lines = parsed
		}
	}

	logs, err := h.customAppSvc.GetLogs(id, lines)
	if err != nil {
		writeCustomAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Data: logs})
}

// handleSetCustomAppAutoStart handles POST /api/custom-apps/{id}/autostart.
func (h *Handler) handleSetCustomAppAutoStart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	app, err := h.customAppSvc.SetAutoStart(id, req.Enabled)
	if err != nil {
		h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_AUTOSTART_CHANGE_FAILED", "failed", err.Error(), "")
		writeCustomAppError(w, err)
		return
	}

	direction := "disabled"
	if req.Enabled {
		direction = "enabled"
	}
	h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_AUTOSTART_CHANGED", "success",
		fmt.Sprintf("%s autostart for %q", direction, app.Name), "")

	writeJSON(w, http.StatusOK, APIResponse{Data: customapp.ToDTO(app)})
}

// handleListEvents handles GET /api/events.
func (h *Handler) handleListEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := events.Filter{
		TargetID:   q.Get("targetId"),
		TargetType: q.Get("targetType"),
		Action:     q.Get("action"),
		Status:     q.Get("status"),
	}
	filter.Limit = parseIntOrDefault(q.Get("limit"), 100)

	list, err := h.eventSvc.List(filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "EVENT_LOG_QUERY_FAILED", "Failed to query events")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Data: list})
}

// handleUpdateCustomApp handles PATCH /api/custom-apps/{id}.
func (h *Handler) handleUpdateCustomApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req customapp.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	app, err := h.customAppSvc.Update(id, &req)
	if err != nil {
		h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_UPDATE_FAILED", "failed", err.Error(), "")
		writeCustomAppError(w, err)
		return
	}

	h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_UPDATED", "success",
		fmt.Sprintf("Updated custom app %q", app.Name), "")

	writeJSON(w, http.StatusOK, APIResponse{Data: customapp.ToDTO(app)})
}

// handleDeleteCustomApp handles DELETE /api/custom-apps/{id}.
func (h *Handler) handleDeleteCustomApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.customAppSvc.Delete(id); err != nil {
		h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_DELETE_FAILED", "failed", err.Error(), "")
		writeCustomAppError(w, err)
		return
	}

	h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_DELETED", "success",
		fmt.Sprintf("Deleted custom app %q", id), "")

	writeJSON(w, http.StatusOK, APIResponse{Data: map[string]string{"status": "deleted"}})
}

// --- Target types for unified API ---

const (
	targetTypeWindowsService = "windows_service"
	targetTypeCustomApp      = "custom_app"
)

// TargetDTO is the unified managed target DTO.
type TargetDTO struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Status         string   `json:"status"`
	AutoStart      bool     `json:"autoStart"`
	ExecutablePath string   `json:"executablePath,omitempty"`
	WorkingDir     string   `json:"workingDir,omitempty"`
	Args           []string `json:"args,omitempty"`
	StartType      string   `json:"startType,omitempty"`
	PID            int      `json:"pid,omitempty"`
	Protected      bool     `json:"protected,omitempty"`
	CanStop        bool     `json:"canStop,omitempty"`
}

// --- Settings handlers ---

// handleGetSettings handles GET /api/settings.
func (h *Handler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse{Data: h.settingsStore.Get()})
}

// handlePatchSettings handles PATCH /api/settings.
func (h *Handler) handlePatchSettings(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	updated, err := h.settingsStore.Update(updates)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SETTINGS_SAVE_FAILED", "Failed to save settings")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Data: updated})
}

// --- Targets unified handlers ---

// handleListTargets handles GET /api/targets.
func (h *Handler) handleListTargets(w http.ResponseWriter, r *http.Request) {
	var targets []TargetDTO
	typeFilter := r.URL.Query().Get("type")

	// Fetch Windows Services
	if typeFilter == "" || typeFilter == targetTypeWindowsService {
		services, err := h.svcProvider.List(r.Context(), windowsservice.Filter{IncludeProtected: false})
		if err == nil {
			for _, svc := range services {
				targets = append(targets, TargetDTO{
					ID:             "windows_service:" + string(svc.ServiceName),
					Name:           svc.DisplayName,
					Type:           targetTypeWindowsService,
					Status:         string(svc.Status),
					AutoStart:      svc.StartType == domain.StartTypeAutomatic || svc.StartType == domain.StartTypeAutomaticDelayed,
					ExecutablePath: svc.ExecutablePath,
					StartType:      string(svc.StartType),
					Protected:      windowsservice.IsProtected(svc.ServiceName),
					CanStop:        svc.CanStop,
				})
			}
		}
	}

	// Fetch Custom Apps
	if typeFilter == "" || typeFilter == targetTypeCustomApp {
		apps, err := h.customAppSvc.List(r.URL.Query().Get("keyword"))
		if err == nil {
			for _, app := range apps {
				targets = append(targets, TargetDTO{
					ID:             app.ID,
					Name:           app.Name,
					Type:           targetTypeCustomApp,
					Status:         string(app.Status),
					AutoStart:      app.AutoStart,
					ExecutablePath: app.ExecutablePath,
					WorkingDir:     app.WorkingDir,
					Args:           app.Args,
					PID:            app.PID,
				})
			}
		}
	}

	writeJSON(w, http.StatusOK, APIResponse{Data: targets})
}

// handleGetTarget handles GET /api/targets/{id}.
func (h *Handler) handleGetTarget(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	targetType, targetID := parseTargetID(id)

	switch targetType {
	case targetTypeWindowsService:
		svc, err := h.svcProvider.Get(r.Context(), targetID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, APIResponse{Data: TargetDTO{
			ID:             id,
			Name:           svc.DisplayName,
			Type:           targetTypeWindowsService,
			Status:         string(svc.Status),
			AutoStart:      svc.StartType == domain.StartTypeAutomatic || svc.StartType == domain.StartTypeAutomaticDelayed,
			ExecutablePath: svc.ExecutablePath,
			StartType:      string(svc.StartType),
			Protected:      windowsservice.IsProtected(svc.ServiceName),
			CanStop:        svc.CanStop,
		}})

	case targetTypeCustomApp:
		app, err := h.customAppSvc.GetByID(targetID)
		if err != nil {
			writeCustomAppError(w, err)
			return
		}
		dto := customapp.ToDTO(app)
		writeJSON(w, http.StatusOK, APIResponse{Data: TargetDTO{
			ID:             id,
			Name:           dto.Name,
			Type:           targetTypeCustomApp,
			Status:         dto.Status,
			AutoStart:      dto.AutoStart,
			ExecutablePath: dto.ExecutablePath,
			WorkingDir:     dto.WorkingDir,
			Args:           dto.Args,
			PID:            dto.PID,
		}})

	default:
		writeError(w, http.StatusBadRequest, "INVALID_TARGET_ID", "Invalid target ID format")
	}
}

// handleStartTarget handles POST /api/targets/{id}/start.
func (h *Handler) handleStartTarget(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	targetType, targetID := parseTargetID(id)

	switch targetType {
	case targetTypeWindowsService:
		if err := h.svcProvider.Start(r.Context(), targetID); err != nil {
			writeServiceError(w, err)
			return
		}
		svc, err := h.svcProvider.Get(r.Context(), targetID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, APIResponse{Data: ServiceActionResponse{
			ServiceName: string(svc.ServiceName),
			Status:      string(svc.Status),
		}})

	case targetTypeCustomApp:
		h.handleStartCustomApp(w, r)

	default:
		writeError(w, http.StatusBadRequest, "INVALID_TARGET_ID", "Invalid target ID format")
	}
}

// handleStopTarget handles POST /api/targets/{id}/stop.
func (h *Handler) handleStopTarget(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	targetType, targetID := parseTargetID(id)

	switch targetType {
	case targetTypeWindowsService:
		if err := h.svcProvider.Stop(r.Context(), targetID); err != nil {
			writeServiceError(w, err)
			return
		}
		svc, err := h.svcProvider.Get(r.Context(), targetID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, APIResponse{Data: ServiceActionResponse{
			ServiceName: string(svc.ServiceName),
			Status:      string(svc.Status),
		}})

	case targetTypeCustomApp:
		h.handleStopCustomApp(w, r)

	default:
		writeError(w, http.StatusBadRequest, "INVALID_TARGET_ID", "Invalid target ID format")
	}
}

// handleRestartTarget handles POST /api/targets/{id}/restart.
func (h *Handler) handleRestartTarget(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	targetType, targetID := parseTargetID(id)

	switch targetType {
	case targetTypeWindowsService:
		if err := h.svcProvider.Restart(r.Context(), targetID); err != nil {
			writeServiceError(w, err)
			return
		}
		svc, err := h.svcProvider.Get(r.Context(), targetID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, APIResponse{Data: ServiceActionResponse{
			ServiceName: string(svc.ServiceName),
			Status:      string(svc.Status),
		}})

	case targetTypeCustomApp:
		h.handleRestartCustomApp(w, r)

	default:
		writeError(w, http.StatusBadRequest, "INVALID_TARGET_ID", "Invalid target ID format")
	}
}

// handleSetTargetAutoStart handles POST /api/targets/{id}/autostart.
func (h *Handler) handleSetTargetAutoStart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	targetType, _ := parseTargetID(id)

	switch targetType {
	case targetTypeCustomApp:
		h.handleSetCustomAppAutoStart(w, r)

	case targetTypeWindowsService:
		writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Use PATCH .../start-type for Windows Services")

	default:
		writeError(w, http.StatusBadRequest, "INVALID_TARGET_ID", "Invalid target ID format")
	}
}

// handleGetTargetLogs handles GET /api/targets/{id}/logs.
func (h *Handler) handleGetTargetLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	targetType, _ := parseTargetID(id)

	switch targetType {
	case targetTypeCustomApp:
		h.handleGetCustomAppLogs(w, r)

	default:
		writeJSON(w, http.StatusOK, APIResponse{Data: map[string]interface{}{
			"stdout": []string{},
			"stderr": []string{},
		}})
	}
}

// parseTargetID splits a target ID like "windows_service:Spooler" or "custom_app:uuid"
// into type and actual ID parts. If no prefix is found, defaults to custom_app.
func parseTargetID(id string) (targetType, targetID string) {
	for _, prefix := range []string{"windows_service:", "custom_app:"} {
		if len(id) > len(prefix) && id[:len(prefix)] == prefix {
			return prefix[:len(prefix)-1], id[len(prefix):]
		}
	}
	return "custom_app", id
}

// handleRestartCustomApp stops and starts a Custom App.
func (h *Handler) handleRestartCustomApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	app, err := h.customAppSvc.Stop(id)
	if err != nil {
		h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_STOP_FAILED", "failed", err.Error(), "")
		writeCustomAppError(w, err)
		return
	}

	app, err = h.customAppSvc.Start(id)
	if err != nil {
		h.eventSvc.Record(id, "custom_app", "CUSTOM_APP_START_FAILED", "failed", err.Error(), "")
		writeCustomAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Data: customapp.ToDTO(app)})
}

// parseInt parses a string to int, returns 0 on failure.
func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number: %s", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// parseIntOrDefault parses a string to int, returns defaultVal on failure.
func parseIntOrDefault(s string, defaultVal int) int {
	n, err := parseInt(s)
	if err != nil {
		return defaultVal
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, APIError{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// writeCustomAppError handles customapp.ServiceError for writing API errors.
func writeCustomAppError(w http.ResponseWriter, err error) {
	// Try customapp.ServiceError first
	var customAppErr *customapp.ServiceError
	if errorsAsCustomApp(err, &customAppErr) {
		status := customAppErrorCodeToHTTPStatus(customAppErr.Code)
		writeError(w, status, customAppErr.Code, customAppErr.Message)
		return
	}

	// Try autostart.ServiceError
	var autoErr *autostart.ServiceError
	if errorsAsAutoStart(err, &autoErr) {
		status := autoStartErrorCodeToHTTPStatus(autoErr.Code)
		writeError(w, status, autoErr.Code, autoErr.Message)
		return
	}

	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
}

// customAppErrorCodeToHTTPStatus maps Custom App error codes to HTTP status codes.
func customAppErrorCodeToHTTPStatus(code string) int {
	switch code {
	case customapp.ErrCodeNotFound:
		return http.StatusNotFound
	case "VALIDATION_ERROR":
		return http.StatusBadRequest
	case customapp.ErrCodeDeleteRunningDenied:
		return http.StatusConflict
	case customapp.ErrCodeAlreadyRunning:
		return http.StatusConflict
	case customapp.ErrCodeNotRunning:
		return http.StatusBadRequest
	case customapp.ErrCodeInvalidExecutable:
		return http.StatusBadRequest
	case customapp.ErrCodeInvalidWorkingDir:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// errorsAsCustomApp unwraps a *customapp.ServiceError.
func errorsAsCustomApp(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	svcErr, ok := err.(*customapp.ServiceError)
	if !ok {
		return false
	}
	switch t := target.(type) {
	case **customapp.ServiceError:
		*t = svcErr
		return true
	}
	return false
}

// writeServiceError unwraps a ServiceError and writes the appropriate HTTP response.
// Non-ServiceError errors are returned as 500.
func writeServiceError(w http.ResponseWriter, err error) {
	var svcErr *windowsservice.ServiceError
	if errorsAs(err, &svcErr) {
		status := errorCodeToHTTPStatus(svcErr.Code)
		writeError(w, status, svcErr.Code, svcErr.Message)
	} else {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
	}
}

// errorCodeToHTTPStatus maps well-known error codes to HTTP status codes.
func errorCodeToHTTPStatus(code string) int {
	switch code {
	case windowsservice.ErrCodeNotFound:
		return http.StatusNotFound
	case windowsservice.ErrCodeProtected:
		return http.StatusForbidden
	case windowsservice.ErrCodePermissionDenied:
		return http.StatusForbidden
	case windowsservice.ErrCodeInvalidStartType:
		return http.StatusBadRequest
	case windowsservice.ErrCodeUnsupported:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

// errorsAs is a helper to unwrap *ServiceError without importing errors everywhere.
func errorsAs(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	svcErr, ok := err.(*windowsservice.ServiceError)
	if !ok {
		return false
	}
	switch t := target.(type) {
	case **windowsservice.ServiceError:
		*t = svcErr
		return true
	}
	return false
}

// autoStartErrorCodeToHTTPStatus maps autostart error codes to HTTP status codes.
func autoStartErrorCodeToHTTPStatus(code string) int {
	switch code {
	case autostart.ErrCodeUnsupported:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

// errorsAsAutoStart unwraps a *autostart.ServiceError.
func errorsAsAutoStart(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	svcErr, ok := err.(*autostart.ServiceError)
	if !ok {
		return false
	}
	switch t := target.(type) {
	case **autostart.ServiceError:
		*t = svcErr
		return true
	}
	return false
}
