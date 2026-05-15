import React, { useEffect, useState, useCallback } from 'react';
import {
  Layout,
  Typography,
  Table,
  Tag,
  Input,
  Space,
  Button,
  Modal,
  Form,
  message,
  Tooltip,
  Row,
  Col,
  Card,
  Switch,
} from 'antd';
import {
  ReloadOutlined,
  PlayCircleOutlined,
  StopOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  FileTextOutlined,
  ExclamationCircleOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import type { CustomAppDTO, CreateCustomAppRequest, UpdateCustomAppRequest } from '../types/api';
import {
  listCustomApps,
  createCustomApp,
  updateCustomApp,
  deleteCustomApp,
  startCustomApp,
  stopCustomApp,
  getCustomAppLogs,
  setCustomAppAutoStart,
} from '../api/client';

const { Content } = Layout;
const { Title, Text } = Typography;

const statusColors: Record<string, string> = {
  running: 'success',
  stopped: 'default',
  error: 'error',
  unknown: 'default',
};

const statusLabels: Record<string, string> = {
  running: 'Running',
  stopped: 'Stopped',
  error: 'Error',
  unknown: 'Unknown',
};

const CustomApps: React.FC = () => {
  const [apps, setApps] = useState<CustomAppDTO[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [formVisible, setFormVisible] = useState(false);
  const [editingApp, setEditingApp] = useState<CustomAppDTO | null>(null);
  const [form] = Form.useForm();
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [logsVisible, setLogsVisible] = useState(false);
  const [logsData, setLogsData] = useState<{ stdout: string[]; stderr: string[] } | null>(null);
  const [logsLoading, setLogsLoading] = useState(false);

  const fetchApps = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listCustomApps(keyword || undefined);
      setApps(res.data || []);
    } catch (err) {
      message.error(`Failed to load apps: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  }, [keyword]);

  useEffect(() => {
    fetchApps();
  }, [fetchApps]);

  const openCreateForm = () => {
    setEditingApp(null);
    form.resetFields();
    setFormVisible(true);
  };

  const openEditForm = (app: CustomAppDTO) => {
    setEditingApp(app);
    form.setFieldsValue({
      name: app.name,
      executablePath: app.executablePath,
      workingDir: app.workingDir || '',
      args: (app.args || []).join('\n'),
      autoStart: app.autoStart,
    });
    setFormVisible(true);
  };

  const handleFormSubmit = async () => {
    try {
      const values = await form.validateFields();
      const body: CreateCustomAppRequest & { id?: string } = {
        name: values.name,
        executablePath: values.executablePath,
        workingDir: values.workingDir || undefined,
        args: values.args ? values.args.split('\n').filter((s: string) => s.trim()) : undefined,
        autoStart: values.autoStart || false,
      };

      if (editingApp) {
        await updateCustomApp(editingApp.id, body as UpdateCustomAppRequest);
        message.success('App updated');
      } else {
        await createCustomApp(body);
        message.success('App created');
      }

      setFormVisible(false);
      setEditingApp(null);
      form.resetFields();
      await fetchApps();
    } catch (err) {
      if (err instanceof Error) {
        message.error(`Failed: ${err.message}`);
      }
      // Validation errors from form are shown automatically
    }
  };

  const handleDelete = (app: CustomAppDTO) => {
    Modal.confirm({
      title: `Delete "${app.name}"?`,
      icon: <ExclamationCircleOutlined />,
      content: (
        <span>
          This will only remove the management configuration.
          {app.status === 'running' && (
            <Text type="danger" style={{ display: 'block', marginTop: 8 }}>
              This app is currently running. Delete will be rejected. Stop it first.
            </Text>
          )}
        </span>
      ),
      okText: 'Delete',
      okType: 'danger',
      cancelText: 'Cancel',
      onOk: async () => {
        setActionLoading(app.id);
        try {
          await deleteCustomApp(app.id);
          message.success(`${app.name} deleted`);
          await fetchApps();
        } catch (err) {
          message.error(`Failed to delete: ${err instanceof Error ? err.message : 'Unknown error'}`);
        } finally {
          setActionLoading(null);
        }
      },
    });
  };

  const handleStart = async (app: CustomAppDTO) => {
    setActionLoading(app.id);
    try {
      await startCustomApp(app.id);
      message.success(`${app.name} started`);
      await fetchApps();
    } catch (err) {
      message.error(`Failed to start: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setActionLoading(null);
    }
  };

  const handleStop = (app: CustomAppDTO) => {
    Modal.confirm({
      title: `Stop "${app.name}"?`,
      icon: <ExclamationCircleOutlined />,
      content: `Are you sure you want to stop "${app.name}"?`,
      okText: 'Stop',
      okType: 'danger',
      cancelText: 'Cancel',
      onOk: async () => {
        setActionLoading(app.id);
        try {
          await stopCustomApp(app.id);
          message.success(`${app.name} stopped`);
          await fetchApps();
        } catch (err) {
          message.error(`Failed to stop: ${err instanceof Error ? err.message : 'Unknown error'}`);
        } finally {
          setActionLoading(null);
        }
      },
    });
  };

  const handleViewLogs = async (app: CustomAppDTO) => {
    setLogsLoading(true);
    setLogsVisible(true);
    try {
      const res = await getCustomAppLogs(app.id, 200);
      setLogsData(res.data);
    } catch (err) {
      message.error(`Failed to load logs: ${err instanceof Error ? err.message : 'Unknown error'}`);
      setLogsVisible(false);
    } finally {
      setLogsLoading(false);
    }
  };

  const handleAutoStartToggle = async (app: CustomAppDTO, enabled: boolean) => {
    try {
      await setCustomAppAutoStart(app.id, enabled);
      message.success(`${app.name} autostart ${enabled ? 'enabled' : 'disabled'}`);
      await fetchApps();
    } catch (err) {
      message.error(`Failed to set autostart: ${err instanceof Error ? err.message : 'Unknown error'}`);
      await fetchApps(); // refresh to revert switch
    }
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      sorter: (a: CustomAppDTO, b: CustomAppDTO) => a.name.localeCompare(b.name),
      width: 200,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={statusColors[status] || 'default'}>
          {statusLabels[status] || status}
        </Tag>
      ),
    },
    {
      title: 'PID',
      dataIndex: 'pid',
      key: 'pid',
      width: 80,
      render: (pid: number | undefined) => pid || '-',
    },
    {
      title: 'Executable Path',
      dataIndex: 'executablePath',
      key: 'executablePath',
      ellipsis: true,
      width: 280,
    },
    {
      title: 'Working Dir',
      dataIndex: 'workingDir',
      key: 'workingDir',
      ellipsis: true,
      width: 180,
      render: (dir: string | undefined) => dir || '-',
    },
    {
      title: 'AutoStart',
      dataIndex: 'autoStart',
      key: 'autoStart',
      width: 90,
      render: (autoStart: boolean, record: CustomAppDTO) => (
        <Switch
          size="small"
          checked={autoStart}
          onChange={(checked) => handleAutoStartToggle(record, checked)}
        />
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 260,
      render: (_: unknown, record: CustomAppDTO) => (
        <Space size="small">
          <Tooltip title="Start">
            <Button
              type="link"
              size="small"
              icon={<PlayCircleOutlined />}
              loading={actionLoading === record.id}
              onClick={() => handleStart(record)}
              disabled={record.status === 'running'}
            />
          </Tooltip>
          <Tooltip title="Stop">
            <Button
              type="link"
              size="small"
              danger
              icon={<StopOutlined />}
              loading={actionLoading === record.id}
              onClick={() => handleStop(record)}
              disabled={record.status !== 'running'}
            />
          </Tooltip>
          <Tooltip title="Edit">
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => openEditForm(record)}
            />
          </Tooltip>
          <Tooltip title="Delete">
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
              loading={actionLoading === record.id}
              onClick={() => handleDelete(record)}
            />
          </Tooltip>
          <Tooltip title="View Logs">
            <Button
              type="link"
              size="small"
              icon={<FileTextOutlined />}
              onClick={() => handleViewLogs(record)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <Layout>
      <Content style={{ padding: '24px' }}>
        <Card>
          <Row gutter={[16, 16]} align="middle" style={{ marginBottom: 16 }}>
            <Col flex="auto">
              <Title level={5} style={{ margin: 0 }}>Custom Apps</Title>
              <Text type="secondary">{apps.length} apps</Text>
            </Col>
            <Col>
              <Space>
                <Input
                  placeholder="Search apps..."
                  prefix={<SearchOutlined />}
                  value={keyword}
                  onChange={(e) => setKeyword(e.target.value)}
                  style={{ width: 200 }}
                  allowClear
                />
                <Button icon={<ReloadOutlined />} onClick={fetchApps} loading={loading}>
                  Refresh
                </Button>
                <Button type="primary" icon={<PlusOutlined />} onClick={openCreateForm}>
                  Add App
                </Button>
              </Space>
            </Col>
          </Row>

          <Table
            dataSource={apps}
            columns={columns}
            rowKey="id"
            loading={loading}
            pagination={{ pageSize: 20, showSizeChanger: true }}
            size="small"
            scroll={{ x: 1200 }}
          />
        </Card>

        {/* Add/Edit Form Modal */}
        <Modal
          title={editingApp ? `Edit "${editingApp.name}"` : 'Add Custom App'}
          open={formVisible}
          onOk={handleFormSubmit}
          onCancel={() => {
            setFormVisible(false);
            setEditingApp(null);
            form.resetFields();
          }}
          okText={editingApp ? 'Save' : 'Create'}
          width={600}
        >
          <Form form={form} layout="vertical">
            <Form.Item name="name" label="Name" rules={[{ required: true, message: 'Name is required' }]}>
              <Input placeholder="My App" />
            </Form.Item>
            <Form.Item
              name="executablePath"
              label="Executable Path"
              rules={[{ required: true, message: 'Executable path is required' }]}
            >
              <Input placeholder="C:\\apps\\my-app\\app.exe" />
            </Form.Item>
            <Form.Item name="workingDir" label="Working Directory">
              <Input placeholder="C:\\apps\\my-app (optional)" />
            </Form.Item>
            <Form.Item name="args" label="Arguments">
              <Input.TextArea
                placeholder="One argument per line&#10;--port&#10;8080"
                rows={3}
              />
            </Form.Item>
            <Form.Item name="autoStart" label="Auto Start" valuePropName="checked">
              <Switch />
            </Form.Item>
          </Form>
        </Modal>

        {/* Logs Modal */}
        <Modal
          title="Application Logs"
          open={logsVisible}
          onCancel={() => {
            setLogsVisible(false);
            setLogsData(null);
          }}
          footer={null}
          width={700}
        >
          {logsLoading ? (
            <Text>Loading logs...</Text>
          ) : logsData ? (
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <Text strong>stdout ({logsData.stdout.length} lines)</Text>
                <pre
                  style={{
                    background: '#f5f5f5',
                    padding: 8,
                    borderRadius: 4,
                    maxHeight: 200,
                    overflow: 'auto',
                    fontSize: 12,
                  }}
                >
                  {logsData.stdout.length > 0
                    ? logsData.stdout.join('\n')
                    : '(empty)'}
                </pre>
              </div>
              <div>
                <Text strong>stderr ({logsData.stderr.length} lines)</Text>
                <pre
                  style={{
                    background: '#fff2f0',
                    padding: 8,
                    borderRadius: 4,
                    maxHeight: 200,
                    overflow: 'auto',
                    fontSize: 12,
                    color: '#cf1322',
                  }}
                >
                  {logsData.stderr.length > 0
                    ? logsData.stderr.join('\n')
                    : '(empty)'}
                </pre>
              </div>
            </Space>
          ) : null}
        </Modal>
      </Content>
    </Layout>
  );
};

export default CustomApps;
