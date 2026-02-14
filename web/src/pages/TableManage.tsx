import { useEffect, useState } from 'react'
import {
  Typography,
  Tabs,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  message,
  Popconfirm,
  Tag,
} from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import type { TableCategory, Table as TableType } from '../api/types'
import {
  listTableCategories,
  createTableCategory,
  updateTableCategory,
  deleteTableCategory,
  listTables,
  createTable,
  getTable,
  updateTable,
  deleteTable,
} from '../api/table'

const TABLE_STATUS = { idle: '空闲', using: '使用中', reserved: '预留', cleaning: '清洁中' } as const

function CategoryTab() {
  const [list, setList] = useState<TableCategory[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form] = Form.useForm()

  const load = async () => {
    setLoading(true)
    try {
      const res = await listTableCategories({ current: page, pageSize: 10 })
      setList(res.categories)
      setTotal(res.total)
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [page])

  const openCreate = () => {
    setEditingId(null)
    form.resetFields()
    setModalOpen(true)
  }

  const openEdit = (r: TableCategory) => {
    setEditingId(r.id)
    form.setFieldsValue({ name: r.name, description: r.description ?? '' })
    setModalOpen(true)
  }

  const onOk = async () => {
    const values = await form.validateFields()
    try {
      if (editingId == null) {
        await createTableCategory({ name: values.name, description: values.description || undefined })
        message.success('创建成功')
      } else {
        await updateTableCategory(editingId, { id: editingId, name: values.name, description: values.description || undefined })
        message.success('更新成功')
      }
      setModalOpen(false)
      load()
    } catch {
      message.error(editingId == null ? '创建失败' : '更新失败')
    }
  }

  const onDelete = async (id: number) => {
    try {
      await deleteTableCategory(id)
      message.success('已删除')
      load()
    } catch {
      message.error('删除失败')
    }
  }

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography.Text type="secondary">餐桌分类，如：大厅、包间。</Typography.Text>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建分类</Button>
      </div>
      <Table
        rowKey="id"
        loading={loading}
        dataSource={list}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 80 },
          { title: '分类名称', dataIndex: 'name' },
          { title: '描述', dataIndex: 'description', ellipsis: true },
          {
            title: '操作',
            width: 120,
            render: (_, r) => (
              <Space>
                <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>编辑</Button>
                <Popconfirm title="确定删除？" onConfirm={() => onDelete(r.id)}>
                  <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
        pagination={{ current: page, pageSize: 10, total, showTotal: (t) => `共 ${t} 条`, onChange: setPage }}
      />
      <Modal title={editingId == null ? '新建分类' : '编辑分类'} open={modalOpen} onOk={onOk} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="分类名称" rules={[{ required: true }]}><Input placeholder="如：大厅" /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={2} placeholder="选填" /></Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

function TableListTab() {
  const [list, setList] = useState<TableType[]>([])
  const [categories, setCategories] = useState<TableCategory[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [categoryFilter, setCategoryFilter] = useState<string | undefined>()
  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form] = Form.useForm()

  const loadCategories = async () => {
    const res = await listTableCategories({ current: 1, pageSize: 200 })
    setCategories(res.categories)
  }

  const loadTables = async () => {
    setLoading(true)
    try {
      const res = await listTables({ current: page, pageSize: 10, category: categoryFilter })
      setList(Array.isArray(res?.tables) ? res.tables : [])
      setTotal(res?.total ?? 0)
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadCategories()
  }, [])

  useEffect(() => {
    loadTables()
  }, [page, categoryFilter])

  const categoryMap = Object.fromEntries(categories.map((c) => [c.id, c.name]))

  const openCreate = () => {
    setEditingId(null)
    form.resetFields()
    form.setFieldsValue({ status: 'idle' })
    setModalOpen(true)
  }

  const openEdit = async (r: TableType) => {
    setEditingId(r.id)
    try {
      const { table } = await getTable(r.id)
      form.setFieldsValue({
        code: table.code,
        status: table.status,
        capacity: table.capacity,
        category_id: table.category_id,
        qr_code: table.qr_code ?? '',
      })
      setModalOpen(true)
    } catch {
      message.error('获取详情失败')
    }
  }

  const onOk = async () => {
    const values = await form.validateFields()
    try {
      if (editingId == null) {
        await createTable({
          code: values.code,
          status: values.status,
          capacity: values.capacity,
          category_id: values.category_id,
          qr_code: values.qr_code || undefined,
        })
        message.success('创建成功')
      } else {
        await updateTable(editingId, {
          id: editingId,
          code: values.code,
          status: values.status,
          capacity: values.capacity,
          category_id: values.category_id,
          qr_code: values.qr_code || undefined,
          create_at: 0,
          update_at: 0,
        })
        message.success('更新成功')
      }
      setModalOpen(false)
      loadTables()
    } catch {
      message.error('操作失败')
    }
  }

  const onDelete = async (id: number) => {
    try {
      await deleteTable(id)
      message.success('已删除')
      loadTables()
    } catch {
      message.error('删除失败')
    }
  }

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 8 }}>
        <Space>
          <Typography.Text type="secondary">按分类：</Typography.Text>
          <Select
            allowClear
            placeholder="全部"
            value={categoryFilter ?? undefined}
            onChange={(v) => setCategoryFilter(v ?? undefined)}
            style={{ minWidth: 120 }}
            options={categories.map((c) => ({ label: c.name, value: c.name }))}
          />
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建餐桌</Button>
      </div>
      <Table
        rowKey="id"
        loading={loading}
        dataSource={list}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 80 },
          { title: '桌号', dataIndex: 'code', width: 100 },
          { title: '状态', dataIndex: 'status', width: 90, render: (s: keyof typeof TABLE_STATUS) => <Tag>{TABLE_STATUS[s] ?? s}</Tag> },
          { title: '可坐人数', dataIndex: 'capacity', width: 90 },
          { title: '分类', dataIndex: 'category_id', width: 100, render: (id: number) => categoryMap[id] ?? id },
          {
            title: '操作',
            width: 120,
            render: (_: unknown, r: TableType) => (
              <Space>
                <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>编辑</Button>
                <Popconfirm title="确定删除？" onConfirm={() => onDelete(r.id)}>
                  <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
        pagination={{ current: page, pageSize: 10, total, showTotal: (t) => `共 ${t} 条`, onChange: setPage }}
      />
      <Modal title={editingId == null ? '新建餐桌' : '编辑餐桌'} open={modalOpen} onOk={onOk} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="code" label="桌号" rules={[{ required: true }]}><Input placeholder="如：A01" /></Form.Item>
          <Form.Item name="status" label="状态" rules={[{ required: true }]}>
            <Select options={Object.entries(TABLE_STATUS).map(([k, v]) => ({ label: v, value: k }))} />
          </Form.Item>
          <Form.Item name="capacity" label="可坐人数" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="category_id" label="所属分类" rules={[{ required: true }]}>
            <Select placeholder="请选择" options={categories.map((c) => ({ label: c.name, value: c.id }))} />
          </Form.Item>
          <Form.Item name="qr_code" label="二维码 URL"><Input placeholder="选填" /></Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default function TableManage() {
  return (
    <div>
      <Typography.Title level={4}>餐桌管理</Typography.Title>
      <Tabs
        defaultActiveKey="category"
        items={[
          { key: 'category', label: '餐桌分类', children: <CategoryTab /> },
          { key: 'table', label: '餐桌列表', children: <TableListTab /> },
        ]}
      />
    </div>
  )
}
