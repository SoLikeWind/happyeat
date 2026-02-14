import { Link, Outlet, useLocation } from 'react-router-dom'
import { Layout, Menu, Dropdown, Space } from 'antd'
import {
  MenuOutlined,
  TableOutlined,
  ShoppingOutlined,
  ShoppingCartOutlined,
  DashboardOutlined,
  UserOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { useAuth } from '../contexts/AuthContext'

const { Header, Sider, Content } = Layout

const navItems = [
  { key: '/menu', icon: <MenuOutlined />, label: '菜单管理' },
  { key: '/tables', icon: <TableOutlined />, label: '餐桌管理' },
  { key: '/order-desk', icon: <ShoppingCartOutlined />, label: '点餐' },
  { key: '/orders', icon: <ShoppingOutlined />, label: '订单管理' },
  { key: '/workbench', icon: <DashboardOutlined />, label: '工作台' },
]

export default function MainLayout() {
  const location = useLocation()
  const { logout } = useAuth()

  const userMenu = {
    items: [
      {
        key: 'logout',
        icon: <LogoutOutlined />,
        label: '退出登录',
        onClick: logout,
      },
    ],
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider theme="dark" breakpoint="lg" collapsedWidth="0">
        <div style={{ height: 32, margin: 16, color: '#fff', fontSize: 18 }}>
          HappyEat
        </div>
        <Menu
          theme="dark"
          selectedKeys={[location.pathname]}
          mode="inline"
          items={navItems.map((item) => ({
            ...item,
            label: <Link to={item.key}>{item.label}</Link>,
          }))}
        />
      </Sider>
      <Layout>
        <Header style={{ padding: '0 24px', background: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'flex-end' }}>
          <Dropdown menu={userMenu} placement="bottomRight">
            <Space style={{ cursor: 'pointer' }}>
              <UserOutlined />
              <span>用户</span>
            </Space>
          </Dropdown>
        </Header>
        <Content style={{ margin: '24px', background: '#fff', padding: 24, minHeight: 280 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
