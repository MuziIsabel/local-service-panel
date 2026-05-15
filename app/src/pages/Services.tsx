import React, { useEffect, useState, useCallback } from 'react';
import {
  Layout,
  Typography,
  Table,
  Tag,
  Input,
  Select,
  Space,
  Button,
  Modal,
  message,
  Tooltip,
  Row,
  Col,
  Card,
} from 'antd';
import {
  ReloadOutlined,
  PlayCircleOutlined,
  StopOutlined,
  SyncOutlined,
  SettingOutlined,
  ExclamationCircleOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import type { WindowsServiceDTO, ServiceStartType } from '../types/api';
import {
  listServices,
  startService,
  stopService,
  restartService,
  setStartType,
} from '../api/client';

const { Content } = Layout;
const { Title, Text } = Typography;

// Color mapping for service status
const statusColors: Record<string, string> = {
  running: 'success',
  stopped: 'default',
  start_pending: 'processing',
  stop_pending: 'warning',
  pause_pending: 'warning',
  paused: 'warning',
  continue_pending: 'processing',
  unknown: 'default',
};

const statusLabels: Record<string, string> = {
  running: 'Running',
  stopped: 'Stopped',
  start_pending: 'Starting...',
  stop_pending: 'Stopping...',
  pause_pending: 'Pausing...',
  paused: 'Paused',
  continue_pending: 'Resuming...',
  unknown: 'Unknown',
};

const startTypeOptions = [
  { value: '', label: 'All Types' },
  { value: 'automatic', label: 'Automatic' },
  { value: 'automatic_delayed', label: 'Auto (Delayed)' },
  { value: 'manual', label: 'Manual' },
  { value: 'disabled', label: 'Disabled' },
];

const statusOptions = [
  { value: '', label: 'All Status' },
  { value: 'running', label: 'Running' },
  { value: 'stopped', label: 'Stopped' },
  { value: 'paused', label: 'Paused' },
  { value: 'unknown', label: 'Unknown' },
];

const startTypeDisplay: Record<string, string> = {
  automatic: 'Automatic',
  automatic_delayed: 'Auto (Delayed)',
  manual: 'Manual',
  disabled: 'Disabled',
  unknown: 'Unknown',
};

const Services: React.FC = () => {
  const [services, setServices] = useState<WindowsServiceDTO[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [startTypeFilter, setStartTypeFilter] = useState('');
  const [startTypeModalVisible, setStartTypeModalVisible] = useState(false);
  const [selectedService, setSelectedService] = useState<WindowsServiceDTO | null>(null);
  const [newStartType, setNewStartType] = useState<ServiceStartType>('manual');
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const fetchServices = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listServices({
        keyword: keyword || undefined,
        status: statusFilter || undefined,
        startType: startTypeFilter || undefined,
      });
      setServices(res.data || []);
    } catch (err) {
      message.error(`Failed to load services: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  }, [keyword, statusFilter, startTypeFilter]);

  useEffect(() => {
    fetchServices();
  }, [fetchServices]);

  const handleAction = async (
    action: () => Promise<unknown>,
    serviceName: string,
    actionLabel: string,
    isDangerous = false,
  ) => {
    if (isDangerous) {
      Modal.confirm({
        title: `Confirm ${actionLabel}`,
        icon: <ExclamationCircleOutlined />,
        content: `Are you sure you want to ${actionLabel.toLowerCase()} "${serviceName}"?`,
        okText: 'Confirm',
        cancelText: 'Cancel',
        onOk: async () => {
          setActionLoading(serviceName);
          try {
            await action();
            message.success(`${serviceName} ${actionLabel.toLowerCase()}ed`);
            await fetchServices();
          } catch (err) {
            message.error(`Failed to ${actionLabel.toLowerCase()} ${serviceName}: ${err instanceof Error ? err.message : 'Unknown error'}`);
          } finally {
            setActionLoading(null);
          }
        },
      });
    } else {
      setActionLoading(serviceName);
      try {
        await action();
        message.success(`${serviceName} ${actionLabel.toLowerCase()}ed`);
        await fetchServices();
      } catch (err) {
        message.error(`Failed to ${actionLabel.toLowerCase()} ${serviceName}: ${err instanceof Error ? err.message : 'Unknown error'}`);
      } finally {
        setActionLoading(null);
      }
    }
  };

  const handleStartTypeChange = async () => {
    if (!selectedService || !newStartType) return;

    setActionLoading(selectedService.serviceName);
    try {
      await setStartType(selectedService.serviceName, newStartType);
      message.success(`Start type for ${selectedService.serviceName} changed to ${startTypeDisplay[newStartType] || newStartType}`);
      setStartTypeModalVisible(false);
      setSelectedService(null);
      await fetchServices();
    } catch (err) {
      message.error(`Failed to change start type: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setActionLoading(null);
    }
  };

  const openStartTypeModal = (svc: WindowsServiceDTO) => {
    setSelectedService(svc);
    // Map current startType to selection
    setNewStartType((svc.startType as ServiceStartType) || 'manual');
    setStartTypeModalVisible(true);
  };

  const columns = [
    {
      title: 'Display Name',
      dataIndex: 'displayName',
      key: 'displayName',
      sorter: (a: WindowsServiceDTO, b: WindowsServiceDTO) => a.displayName.localeCompare(b.displayName),
      width: 250,
    },
    {
      title: 'Service Name',
      dataIndex: 'serviceName',
      key: 'serviceName',
      width: 180,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 130,
      render: (status: string) => (
        <Tag color={statusColors[status] || 'default'}>
          {statusLabels[status] || status}
        </Tag>
      ),
      filters: [
        { text: 'Running', value: 'running' },
        { text: 'Stopped', value: 'stopped' },
        { text: 'Paused', value: 'paused' },
      ],
      onFilter: (value: boolean | React.Key, record: WindowsServiceDTO) => record.status === value,
    },
    {
      title: 'Start Type',
      dataIndex: 'startType',
      key: 'startType',
      width: 140,
      render: (startType: string) => (
        <Text>{startTypeDisplay[startType] || startType}</Text>
      ),
    },
    {
      title: 'Protected',
      dataIndex: 'protected',
      key: 'protected',
      width: 100,
      render: (isProtected: boolean) =>
        isProtected ? (
          <Tag color="red">Protected</Tag>
        ) : (
          <Tag style={{ visibility: 'hidden' }}>-</Tag>
        ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 320,
      render: (_: unknown, record: WindowsServiceDTO) => (
        <Space size="small">
          <Tooltip title="Start">
            <Button
              type="link"
              size="small"
              icon={<PlayCircleOutlined />}
              loading={actionLoading === record.serviceName}
              onClick={() =>
                handleAction(
                  () => startService(record.serviceName),
                  record.serviceName,
                  'Start',
                )
              }
              disabled={record.status === 'running'}
            />
          </Tooltip>
          <Tooltip title="Stop">
            <Button
              type="link"
              size="small"
              danger
              icon={<StopOutlined />}
              loading={actionLoading === record.serviceName}
              onClick={() =>
                handleAction(
                  () => stopService(record.serviceName),
                  record.serviceName,
                  'Stop',
                  !record.protected, // confirm for non-protected, protected will 403 anyway
                )
              }
              disabled={record.status !== 'running' || record.protected}
            />
          </Tooltip>
          <Tooltip title="Restart">
            <Button
              type="link"
              size="small"
              icon={<SyncOutlined />}
              loading={actionLoading === record.serviceName}
              onClick={() =>
                handleAction(
                  () => restartService(record.serviceName),
                  record.serviceName,
                  'Restart',
                  true, // always confirm restart
                )
              }
              disabled={record.status !== 'running' || record.protected}
            />
          </Tooltip>
          <Tooltip title="Change Start Type">
            <Button
              type="link"
              size="small"
              icon={<SettingOutlined />}
              onClick={() => openStartTypeModal(record)}
              disabled={record.protected}
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
              <Title level={5} style={{ margin: 0 }}>
                Windows Services
              </Title>
              <Text type="secondary">{services.length} services found</Text>
            </Col>
            <Col>
              <Space>
                <Input
                  placeholder="Search services..."
                  prefix={<SearchOutlined />}
                  value={keyword}
                  onChange={(e) => setKeyword(e.target.value)}
                  style={{ width: 200 }}
                  allowClear
                />
                <Select
                  value={statusFilter}
                  onChange={setStatusFilter}
                  options={statusOptions}
                  style={{ width: 130 }}
                />
                <Select
                  value={startTypeFilter}
                  onChange={setStartTypeFilter}
                  options={startTypeOptions}
                  style={{ width: 150 }}
                />
                <Button icon={<ReloadOutlined />} onClick={fetchServices} loading={loading}>
                  Refresh
                </Button>
              </Space>
            </Col>
          </Row>
          <Table
            dataSource={services}
            columns={columns}
            rowKey="serviceName"
            loading={loading}
            pagination={{ pageSize: 50, showSizeChanger: true, showTotal: (total) => `${total} services` }}
            size="small"
            scroll={{ x: 1100 }}
          />
        </Card>

        {/* Start Type Change Modal */}
        <Modal
          title={`Change Start Type - ${selectedService?.displayName || ''}`}
          open={startTypeModalVisible}
          onOk={handleStartTypeChange}
          onCancel={() => {
            setStartTypeModalVisible(false);
            setSelectedService(null);
          }}
          confirmLoading={!!actionLoading}
        >
          <Space direction="vertical" style={{ width: '100%' }}>
            <Text>Service: <strong>{selectedService?.serviceName}</strong></Text>
            <Text>Current type: {selectedService ? startTypeDisplay[selectedService.startType] || selectedService.startType : ''}</Text>
            <Select
              value={newStartType}
              onChange={(val) => setNewStartType(val as ServiceStartType)}
              options={[
                { value: 'automatic', label: 'Automatic' },
                { value: 'automatic_delayed', label: 'Automatic (Delayed Start)' },
                { value: 'manual', label: 'Manual' },
                { value: 'disabled', label: 'Disabled' },
              ]}
              style={{ width: '100%' }}
            />
            {newStartType === 'disabled' && (
              <Text type="danger">
                <ExclamationCircleOutlined /> Disabling a service may prevent certain system features from working.
              </Text>
            )}
          </Space>
        </Modal>
      </Content>
    </Layout>
  );
};

export default Services;
