import { Link } from 'react-router-dom'
import { Typography, Card, Row, Col, Space } from 'antd'
import { MenuOutlined, TableOutlined, ShoppingOutlined, ShoppingCartOutlined, DashboardOutlined } from '@ant-design/icons'

const links = [
  { path: '/menu', title: '菜单管理', icon: <MenuOutlined /> },
  { path: '/tables', title: '餐桌管理', icon: <TableOutlined /> },
  { path: '/order-desk', title: '点餐', icon: <ShoppingCartOutlined /> },
  { path: '/orders', title: '订单管理', icon: <ShoppingOutlined /> },
  { path: '/workbench', title: '工作台', icon: <DashboardOutlined /> },
]

export default function Home() {
  return (
    <div>
      <Typography.Title level={4}>HappyEat 管理后台</Typography.Title>
      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        {links.map((item) => (
          <Col key={item.path} xs={24} sm={12} md={6}>
            <Link to={item.path}>
              <Card hoverable>
                <Space>
                  {item.icon}
                  <span>{item.title}</span>
                </Space>
              </Card>
            </Link>
          </Col>
        ))}
      </Row>
    </div>
  )
}
