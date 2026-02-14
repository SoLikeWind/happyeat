import { useEffect, useState } from 'react'
import {
  Typography,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  message,
  Tag,
  Descriptions,
} from 'antd'
import { PlusOutlined, EyeOutlined } from '@ant-design/icons'
import type { Order, Menu, Table as TTable } from '../api/types'
import { listOrders, createOrder, getOrder, updateOrderStatus } from '../api/order'
import { listMenus } from '../api/menu'
import { listTables } from '../api/table'

const ORDER_TYPE_MAP: Record<string, string> = { dine_in: '堂食', takeaway: '打包外带' }
const STATUS_MAP: Record<string, string> = {
  created: '待支付',
  paid: '已支付',
  preparing: '制作中',
  completed: '已完成',
  cancelled: '已取消',
}

export default function OrderManage() {
  const [orders, setOrders] = useState<Order[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState<string | undefined>()
  const [orderTypeFilter, setOrderTypeFilter] = useState<string | undefined>()
  const [modalOpen, setModalOpen] = useState(false)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailOrder, setDetailOrder] = useState<Order | null>(null)
  const [menus, setMenus] = useState<Menu[]>([])
  const [tables, setTables] = useState<TTable[]>([])
  const [form] = Form.useForm()

  const openDetail = async (id: number) => {
    try {
      const { order } = await getOrder(id)
      setDetailOrder(order)
      setDetailOpen(true)
    } catch {
      message.error('获取订单详情失败')
    }
  }

  const handleUpdateStatus = async (id: number, status: string) => {
    try {
      await updateOrderStatus(id, status)
      message.success('状态已更新')
      load()
      if (detailOrder?.id === id) {
        setDetailOrder((o) => (o ? { ...o, status } : null))
      }
    } catch {
      message.error('更新失败')
    }
  }

  const load = async () => {
    setLoading(true)
    try {
      const res = await listOrders({
        current: page,
        pageSize: 10,
        status: statusFilter,
        order_type: orderTypeFilter,
      })
      setOrders(Array.isArray(res?.orders) ? res.orders : [])
      setTotal(res?.total ?? 0)
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [page, statusFilter, orderTypeFilter])

  const openCreate = async () => {
    form.resetFields()
    form.setFieldsValue({ order_type: 'dine_in', items: [{}] })
    try {
      const [menuRes, tableRes] = await Promise.all([
        listMenus({ current: 1, pageSize: 500 }),
        listTables({ current: 1, pageSize: 500 }),
      ])
      setMenus(Array.isArray(menuRes?.menus) ? menuRes.menus : [])
      const tbls = Array.isArray(tableRes?.tables) ? tableRes.tables : []
      setTables(tbls.filter((t) => t.status === 'idle' || t.status === 'using'))
      setModalOpen(true)
    } catch {
      message.error('加载菜单/餐桌失败')
    }
  }

  const onOk = async () => {
    const values = await form.validateFields()
    const orderType = values.order_type
    const items = (values.items ?? []).filter((i: { menu_id?: number; quantity?: number }) => i?.menu_id != null && (i?.quantity ?? 0) > 0)
    if (items.length === 0) {
      message.error('请至少添加一道菜品')
      return
    }
    const menuMap = Object.fromEntries(menus.map((m) => [m.id, m]))
    let totalAmount = 0
    const bodyItems = items.map((i: { menu_id: number; quantity: number; spec_info?: string }) => {
      const menu = menuMap[i.menu_id]
      const unitPrice = menu ? menu.price : 0
      const amount = unitPrice * i.quantity
      totalAmount += amount
      return {
        menu_name: menu?.name ?? '',
        quantity: i.quantity,
        unit_price: unitPrice,
        spec_info: i.spec_info || undefined,
      }
    })
    try {
      await createOrder({
        order_type: orderType,
        table_id: orderType === 'dine_in' ? values.table_id : undefined,
        items: bodyItems,
        total_amount: Math.round(totalAmount * 100) / 100,
        remark: values.remark || undefined,
      })
      message.success('创建成功')
      setModalOpen(false)
      load()
    } catch {
      message.error('创建失败')
    }
  }

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 8 }}>
        <Space wrap>
          <Typography.Text type="secondary">状态：</Typography.Text>
          <Select
            allowClear
            placeholder="全部"
            value={statusFilter ?? undefined}
            onChange={setStatusFilter}
            style={{ minWidth: 100 }}
            options={Object.entries(STATUS_MAP).map(([k, v]) => ({ label: v, value: k }))}
          />
          <Typography.Text type="secondary">类型：</Typography.Text>
          <Select
            allowClear
            placeholder="全部"
            value={orderTypeFilter ?? undefined}
            onChange={setOrderTypeFilter}
            style={{ minWidth: 110 }}
            options={Object.entries(ORDER_TYPE_MAP).map(([k, v]) => ({ label: v, value: k }))}
          />
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建订单</Button>
      </div>
      <Table
        rowKey="id"
        loading={loading}
        dataSource={orders}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 70 },
          { title: '订单号', dataIndex: 'order_no', width: 140, ellipsis: true },
          { title: '类型', dataIndex: 'order_type', width: 90, render: (t: string) => ORDER_TYPE_MAP[t] ?? t },
          { title: '状态', dataIndex: 'status', width: 90, render: (s: string) => <Tag>{STATUS_MAP[s] ?? s}</Tag> },
          { title: '金额', dataIndex: 'total_amount', width: 90, render: (v: number) => `¥${v?.toFixed(2) ?? '0.00'}` },
          { title: '桌台', dataIndex: 'table_id', width: 70, render: (v: number) => (v ? `#${v}` : '-') },
          { title: '备注', dataIndex: 'remark', ellipsis: true },
          {
            title: '操作',
            width: 180,
            fixed: 'right',
            render: (_: unknown, r: Order) => (
              <Space wrap>
                <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => openDetail(r.id)}>
                  查看
                </Button>
                <Select
                  size="small"
                  value={r.status}
                  onChange={(status) => handleUpdateStatus(r.id, status)}
                  style={{ width: 100 }}
                  options={Object.entries(STATUS_MAP).map(([k, v]) => ({ label: v, value: k }))}
                />
              </Space>
            ),
          },
        ]}
        pagination={{ current: page, pageSize: 10, total, showTotal: (t) => `共 ${t} 条`, onChange: setPage }}
      />
      <Modal
        title="新建订单"
        open={modalOpen}
        onOk={onOk}
        onCancel={() => setModalOpen(false)}
        width={560}
        destroyOnClose
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="order_type" label="订单类型" rules={[{ required: true }]}>
            <Select options={Object.entries(ORDER_TYPE_MAP).map(([k, v]) => ({ label: v, value: k }))} />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, cur) => prev?.order_type !== cur?.order_type}>
            {({ getFieldValue }) =>
              getFieldValue('order_type') === 'dine_in' ? (
                <Form.Item name="table_id" label="选择餐桌（选填）">
                  <Select
                    placeholder="请选择餐桌"
                    allowClear
                    options={tables.map((t) => ({ label: `${t.code}（${t.capacity}人）`, value: t.id }))}
                  />
                </Form.Item>
              ) : null
            }
          </Form.Item>
          <Form.Item label="菜品">
            <Form.List name="items">
              {(fields, { add, remove }) => (
                <>
                  {fields.map(({ key, name, ...rest }) => (
                    <Space key={key} style={{ marginBottom: 8 }} align="baseline">
                      <Form.Item {...rest} name={[name, 'menu_id']} rules={[{ required: true }]} noStyle>
                        <Select
                          placeholder="选择菜品"
                          style={{ width: 160 }}
                          options={menus.map((m) => ({ label: `${m.name} ¥${m.price}`, value: m.id }))}
                        />
                      </Form.Item>
                      <Form.Item {...rest} name={[name, 'quantity']} rules={[{ required: true }]} initialValue={1} noStyle>
                        <InputNumber min={1} placeholder="数量" style={{ width: 70 }} />
                      </Form.Item>
                      <Form.Item {...rest} name={[name, 'spec_info']} noStyle>
                        <Input placeholder="规格备注" style={{ width: 100 }} />
                      </Form.Item>
                      <Button type="text" danger onClick={() => remove(name)}>删</Button>
                    </Space>
                  ))}
                  <Button type="dashed" onClick={() => add()} block>+ 添加菜品</Button>
                </>
              )}
            </Form.List>
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea rows={2} placeholder="选填" /></Form.Item>
        </Form>
      </Modal>

      <Modal
        title="订单详情"
        open={detailOpen}
        onCancel={() => setDetailOpen(false)}
        footer={null}
        width={560}
      >
        {detailOrder && (
          <>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="订单号">{detailOrder.order_no}</Descriptions.Item>
              <Descriptions.Item label="类型">{ORDER_TYPE_MAP[detailOrder.order_type] ?? detailOrder.order_type}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Select
                  size="small"
                  value={detailOrder.status}
                  onChange={(status) => handleUpdateStatus(detailOrder.id, status)}
                  style={{ width: 120 }}
                  options={Object.entries(STATUS_MAP).map(([k, v]) => ({ label: v, value: k }))}
                />
              </Descriptions.Item>
              <Descriptions.Item label="金额">¥{detailOrder.total_amount?.toFixed(2) ?? '0.00'}</Descriptions.Item>
              <Descriptions.Item label="桌台">{detailOrder.table_id ? `#${detailOrder.table_id}` : '-'}</Descriptions.Item>
              <Descriptions.Item label="备注">{detailOrder.remark || '-'}</Descriptions.Item>
            </Descriptions>
            <Typography.Title level={5} style={{ marginTop: 16, marginBottom: 8 }}>订单明细</Typography.Title>
            <Table
              size="small"
              rowKey={(_, i) => String(i)}
              dataSource={detailOrder.items ?? []}
              pagination={false}
              columns={[
                { title: '菜品', dataIndex: 'menu_name' },
                { title: '数量', dataIndex: 'quantity', width: 70 },
                { title: '单价', dataIndex: 'unit_price', width: 90, render: (v: number) => `¥${v?.toFixed(2) ?? '0.00'}` },
                { title: '小计', dataIndex: 'amount', width: 90, render: (v: number) => `¥${v?.toFixed(2) ?? '0.00'}` },
                { title: '规格', dataIndex: 'spec_info', ellipsis: true },
              ]}
            />
          </>
        )}
      </Modal>
    </div>
  )
}
