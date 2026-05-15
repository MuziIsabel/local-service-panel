import React, { useEffect, useState, useCallback } from 'react';
import { Card, Typography, Tag, Space, Spin, Row, Col, Statistic, Table, Alert } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  LoadingOutlined,
  AppstoreOutlined,
  CodeOutlined,
  ThunderboltOutlined,
  WarningOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import type { AgentStatus, WindowsServiceDTO, CustomAppDTO, EventLogDTO } from '../types/api';
import { checkHealthz, listServices, listCustomApps, listEvents } from '../api/client';

const { Title, Text } = Typography;

interface DashboardData {
  services: WindowsServiceDTO[];
  customApps: CustomAppDTO[];
  events: EventLogDTO[];
}

const Dashboard: React.FC = () => {
  const [status, setStatus] = useState<AgentStatus>('checking');
  const [version, setVersion] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [healthzRes, servicesRes, appsRes, eventsRes] = await Promise.allSettled([
        checkHealthz(),
        listServices(),
        listCustomApps(),
        listEvents({ limit: 10 }),
      ]);

      // Handle healthz
      if (healthzRes.status === 'fulfilled' && healthzRes.value.data?.status === 'ok') {
        setStatus('connected');
        setVersion(healthzRes.value.data.version);
        setError('');
      } else {
        setStatus('disconnected');
        setError('Cannot connect to Agent');
      }

      // Handle services, apps, events
      const services = servicesRes.status === 'fulfilled' ? servicesRes.value.data : [];
      const customApps = appsRes.status === 'fulfilled' ? appsRes.value.data : [];
      const events = eventsRes.status === 'fulfilled' ? eventsRes.value.data : [];

      setData({ services, customApps, events });
    } catch {
      setStatus('disconnected');
      setError('Cannot connect to Agent');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000);
    return () => clearInterval(interval);
  }, [fetchData]);

  const runningServices = data?.services.filter(s => s.status === 'running').length ?? 0;
  const runningApps = data?.customApps.filter(a => a.status === 'running').length ?? 0;
  const failedEvents = data?.events.filter(e => e.status === 'failed') ?? [];
  const isConnected = status === 'connected';

  const eventColumns = [
    {
      title: 'Time',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (t: string) => {
        const d = new Date(t);
        return d.toLocaleString();
      },
    },
    {
      title: 'Action',
      dataIndex: 'action',
      key: 'action',
      width: 280,
      render: (action: string) => {
        const colorMap: Record<string, string> = {
          CUSTOM_APP_STARTED: 'green',
          CUSTOM_APP_STOPPED: 'orange',
          CUSTOM_APP_CREATED: 'blue',
          CUSTOM_APP_DELETED: 'red',
          CUSTOM_APP_UPDATED: 'cyan',
          CUSTOM_APP_AUTOSTART_CHANGED: 'purple',
          WINDOWS_SERVICE_STARTED: 'green',
          WINDOWS_SERVICE_STOPPED: 'orange',
          WINDOWS_SERVICE_RESTARTED: 'blue',
          WINDOWS_SERVICE_START_TYPE_CHANGED: 'purple',
        };
        const failedActions = [
          'CUSTOM_APP_START_FAILED',
          'CUSTOM_APP_STOP_FAILED',
          'CUSTOM_APP_CREATE_FAILED',
          'CUSTOM_APP_DELETE_FAILED',
          'CUSTOM_APP_UPDATE_FAILED',
          'CUSTOM_APP_AUTOSTART_CHANGE_FAILED',
          'WINDOWS_SERVICE_START_FAILED',
          'WINDOWS_SERVICE_STOP_FAILED',
          'WINDOWS_SERVICE_RESTART_FAILED',
          'WINDOWS_SERVICE_START_TYPE_CHANGE_FAILED',
        ];
        if (failedActions.includes(action)) {
          return <Tag color="error">{action}</Tag>;
        }
        return <Tag color={colorMap[action] || 'default'}>{action}</Tag>;
      },
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (s: string) => (
        <Tag color={s === 'success' ? 'success' : s === 'failed' ? 'error' : 'default'}>{s}</Tag>
      ),
    },
    {
      title: 'Message',
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Title level={4} style={{ marginBottom: '16px' }}>Dashboard</Title>

      {/* Connection status */}
      <Card style={{ marginBottom: '16px' }}>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Space>
            <Title level={5} style={{ margin: 0 }}>Agent Status</Title>
            {status === 'checking' && (
              <Tag icon={<Spin indicator={<LoadingOutlined style={{ fontSize: '12px' }} spin />} />} color="processing">
                Checking...
              </Tag>
            )}
            {status === 'connected' && (
              <Tag icon={<CheckCircleOutlined />} color="success">
                Connected
              </Tag>
            )}
            {status === 'disconnected' && (
              <Tag icon={<CloseCircleOutlined />} color="error">
                Disconnected
              </Tag>
            )}
            {version && <Text type="secondary">v{version}</Text>}
          </Space>
          {error && (
            <Alert
              message={error}
              description="Please ensure the Agent is running on 127.0.0.1:17645."
              type="error"
              showIcon
            />
          )}
        </Space>
      </Card>

      {/* Statistics cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: '16px' }}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title="Windows Services"
              value={data?.services.length ?? '-'}
              suffix={
                <Text type="secondary" style={{ fontSize: '14px' }}>
                  / {runningServices} running
                </Text>
              }
              prefix={<AppstoreOutlined style={{ color: '#1677ff' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title="Custom Apps"
              value={data?.customApps.length ?? '-'}
              suffix={
                <Text type="secondary" style={{ fontSize: '14px' }}>
                  / {runningApps} running
                </Text>
              }
              prefix={<CodeOutlined style={{ color: '#52c41a' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title="Agent Uptime"
              value={isConnected ? 'Online' : 'Offline'}
              valueStyle={{ color: isConnected ? '#52c41a' : '#ff4d4f', fontSize: '20px' }}
              prefix={isConnected ? <ThunderboltOutlined /> : <CloseCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title="Recent Errors"
              value={failedEvents.length}
              valueStyle={{ color: failedEvents.length > 0 ? '#ff4d4f' : '#52c41a' }}
              prefix={<WarningOutlined />}
              suffix={
                failedEvents.length > 0 ? (
                  <Text type="danger" style={{ fontSize: '14px' }}>
                    in last 10 events
                  </Text>
                ) : undefined
              }
            />
          </Card>
        </Col>
      </Row>

      {/* Recent Events */}
      <Card
        title={
          <Space>
            <InfoCircleOutlined />
            <span>Recent Events</span>
          </Space>
        }
        loading={loading}
      >
        {data?.events.length ? (
          <Table
            dataSource={data.events}
            columns={eventColumns}
            rowKey="id"
            pagination={false}
            size="small"
          />
        ) : (
          <Text type="secondary">No events recorded yet.</Text>
        )}
      </Card>
    </div>
  );
};

export default Dashboard;
