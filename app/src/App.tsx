import React, { useState } from 'react';
import { ConfigProvider, Layout, Menu, Typography } from 'antd';
import {
  DashboardOutlined,
  AppstoreOutlined,
  CodeOutlined,
} from '@ant-design/icons';
import Dashboard from './pages/Dashboard';
import Services from './pages/Services';
import CustomApps from './pages/CustomApps';

const { Sider, Content } = Layout;
const { Text } = Typography;

// UI version - sync with package.json
const UI_VERSION = '0.6.0';

const App: React.FC = () => {
  const [page, setPage] = useState('dashboard');

  const menuItems = [
    {
      key: 'dashboard',
      icon: <DashboardOutlined />,
      label: 'Dashboard',
    },
    {
      key: 'services',
      icon: <AppstoreOutlined />,
      label: 'Services',
    },
    {
      key: 'custom-apps',
      icon: <CodeOutlined />,
      label: 'Custom Apps',
    },
  ];

  const renderPage = () => {
    switch (page) {
      case 'services':
        return <Services />;
      case 'custom-apps':
        return <CustomApps />;
      case 'dashboard':
      default:
        return <Dashboard />;
    }
  };

  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#1677ff',
        },
      }}
    >
      <Layout style={{ minHeight: '100vh' }}>
        <Sider width={200} theme="dark" style={{ display: 'flex', flexDirection: 'column' }}>
          <div style={{ height: 32, margin: 16, color: '#fff', fontWeight: 'bold', fontSize: 16 }}>
            LSP
          </div>
          <Menu
            theme="dark"
            mode="inline"
            selectedKeys={[page]}
            items={menuItems}
            onClick={({ key }) => setPage(key)}
            style={{ flex: 1 }}
          />
          <div style={{ padding: '12px 16px', borderTop: '1px solid rgba(255,255,255,0.1)' }}>
            <Text style={{ color: 'rgba(255,255,255,0.45)', fontSize: 12 }}>
              UI v{UI_VERSION}
            </Text>
          </div>
        </Sider>
        <Layout>
          <Content>{renderPage()}</Content>
        </Layout>
      </Layout>
    </ConfigProvider>
  );
};

export default App;
