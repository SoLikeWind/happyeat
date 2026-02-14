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
  message,
  Popconfirm,
} from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import type { MenuCategory, Menu, MenuSpec } from '../api/types'
import {
  listMenuCategories,
  createMenuCategory,
  updateMenuCategory,
  deleteMenuCategory,
  listMenus,
  createMenu,
  updateMenu,
  deleteMenu,
  getMenu,
} from '../api/menu'

// ============ 第一 Tab：菜单分类 ============

function CategoryTab() {
  const [list, setList] = useState<MenuCategory[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const pageSize = 10

  /** 弹窗：新建 vs 编辑。open 时 id 为空表示新建，有值表示编辑 */
  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form] = Form.useForm()

  const load = async () => {
    setLoading(true)
    try {
      const res = await listMenuCategories({ current: page, pageSize })
      setList(res.categories)
      setTotal(res.total)
    } catch (e) {
      message.error('加载分类列表失败')
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

  const openEdit = (record: MenuCategory) => {
    setEditingId(record.id)
    form.setFieldsValue({ name: record.name, description: record.description ?? '' })
    setModalOpen(true)
  }

  const onModalOk = async () => {
    const values = await form.validateFields()
    try {
      if (editingId == null) {
        await createMenuCategory({ name: values.name, description: values.description || undefined })
        message.success('创建成功')
      } else {
        await updateMenuCategory(editingId, {
          id: editingId,
          name: values.name,
          description: values.description || undefined,
        })
        message.success('更新成功')
      }
      setModalOpen(false)
      load()
    } catch (e) {
      message.error(editingId == null ? '创建失败' : '更新失败')
    }
  }

  const onDelete = async (id: number) => {
    try {
      await deleteMenuCategory(id)
      message.success('已删除')
      load()
    } catch (e) {
      message.error('删除失败')
    }
  }

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography.Text type="secondary">管理菜品分类，如：热菜、凉菜、饮品。</Typography.Text>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          新建分类
        </Button>
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
            render: (_, record) => (
              <Space>
                <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEdit(record)}>
                  编辑
                </Button>
                <Popconfirm title="确定删除该分类？" onConfirm={() => onDelete(record.id)}>
                  <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                    删除
                  </Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: false,
          showTotal: (t) => `共 ${t} 条`,
          onChange: setPage,
        }}
      />
      <Modal
        title={editingId == null ? '新建分类' : '编辑分类'}
        open={modalOpen}
        onOk={onModalOk}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="分类名称" rules={[{ required: true, message: '请输入分类名称' }]}>
            <Input placeholder="如：热菜" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="选填" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

// ============ 第二 Tab：菜品列表 ============

function MenuListTab() {
  const [menus, setMenus] = useState<Menu[]>([])
  const [categories, setCategories] = useState<MenuCategory[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [categoryFilter, setCategoryFilter] = useState<string | undefined>(undefined)
  const pageSize = 10

  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form] = Form.useForm()

  const loadCategories = async () => {
    const res = await listMenuCategories({ current: 1, pageSize: 200 })
    setCategories(res.categories)
  }

  const loadMenus = async () => {
    setLoading(true)
    try {
      const res = await listMenus({
        current: page,
        pageSize,
        category: categoryFilter,
      })
      setMenus(res.menus)
      setTotal(res.total)
    } catch (e) {
      message.error('加载菜品列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadCategories()
  }, [])

  useEffect(() => {
    loadMenus()
  }, [page, categoryFilter])

  const categoryMap = Object.fromEntries(categories.map((c) => [c.id, c.name]))

  const openCreate = () => {
    setEditingId(null)
    form.resetFields()
    form.setFieldValue('specs', [])
    setModalOpen(true)
  }

  const openEdit = async (record: Menu) => {
    setEditingId(record.id)
    try {
      const { menu } = await getMenu(record.id)
      form.setFieldsValue({
        name: menu.name,
        price: menu.price,
        category_id: menu.category_id,
        description: menu.description ?? '',
        image: menu.image ?? '',
        specs: (menu.specs && menu.specs.length) ? menu.specs : [],
      })
      setModalOpen(true)
    } catch (e) {
      message.error('获取菜品详情失败')
    }
  }

  const onModalOk = async () => {
    const values = await form.validateFields()
    const specs: MenuSpec[] = (values.specs ?? [])
      .filter((s: { spec_type?: string }) => s?.spec_type)
      .map((s: { spec_type: string; spec_value: string; price_delta: number }) => ({
        spec_type: s.spec_type,
        spec_value: s.spec_value,
        price_delta: Number(s.price_delta) || 0,
      }))
    try {
      if (editingId == null) {
        await createMenu({
          name: values.name,
          price: Number(values.price),
          category_id: Number(values.category_id),
          description: values.description || undefined,
          image: values.image || undefined,
          specs: specs.length ? specs : undefined,
        })
        message.success('创建成功')
      } else {
        const menu: Menu = {
          id: editingId,
          name: values.name,
          price: Number(values.price),
          category_id: Number(values.category_id),
          description: values.description || undefined,
          image: values.image || undefined,
          specs: specs.length ? specs : undefined,
          create_at: 0,
          update_at: 0,
        }
        await updateMenu(editingId, menu)
        message.success('更新成功')
      }
      setModalOpen(false)
      loadMenus()
    } catch (e) {
      message.error(editingId == null ? '创建失败' : '更新失败')
    }
  }

  const onDelete = async (id: number) => {
    try {
      await deleteMenu(id)
      message.success('已删除')
      loadMenus()
    } catch (e) {
      message.error('删除失败')
    }
  }

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 8 }}>
        <Space>
          <Typography.Text type="secondary">按分类筛选：</Typography.Text>
          <Select
            allowClear
            placeholder="全部"
            value={categoryFilter ?? undefined}
            onChange={(v) => setCategoryFilter(v ?? undefined)}
            style={{ minWidth: 120 }}
            options={categories.map((c) => ({ label: c.name, value: c.name }))}
          />
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          新建菜品
        </Button>
      </div>
      <Table
        rowKey="id"
        loading={loading}
        dataSource={menus}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 80 },
          {
            title: '图片',
            dataIndex: 'image',
            width: 72,
            render: (url: string | undefined) =>
              url ? (
                <img
                  src={url}
                  alt=""
                  style={{ width: 48, height: 48, objectFit: 'cover', borderRadius: 4 }}
                  onError={(e) => {
                    (e.target as HTMLImageElement).style.display = 'none'
                  }}
                />
              ) : (
                <span style={{ color: '#999', fontSize: 12 }}>无图</span>
              ),
          },
          { title: '菜品名称', dataIndex: 'name' },
          { title: '价格', dataIndex: 'price', width: 80, render: (v: number) => `¥${v.toFixed(2)}` },
          {
            title: '分类',
            dataIndex: 'category_id',
            width: 100,
            render: (id: number) => categoryMap[id] ?? id,
          },
          { title: '描述', dataIndex: 'description', ellipsis: true },
          {
            title: '操作',
            width: 120,
            render: (_, record) => (
              <Space>
                <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEdit(record)}>
                  编辑
                </Button>
                <Popconfirm title="确定删除该菜品？" onConfirm={() => onDelete(record.id)}>
                  <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                    删除
                  </Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: false,
          showTotal: (t) => `共 ${t} 条`,
          onChange: setPage,
        }}
      />
      <Modal
        title={editingId == null ? '新建菜品' : '编辑菜品'}
        open={modalOpen}
        onOk={onModalOk}
        onCancel={() => setModalOpen(false)}
        width={560}
        destroyOnClose
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="菜品名称" rules={[{ required: true }]}>
            <Input placeholder="如：宫保鸡丁" />
          </Form.Item>
          <Form.Item name="price" label="价格（元）" rules={[{ required: true }]}>
            <Input type="number" step={0.01} min={0} placeholder="0.00" />
          </Form.Item>
          <Form.Item name="category_id" label="所属分类" rules={[{ required: true, message: '请选择分类' }]}>
            <Select
              placeholder="请选择分类"
              options={categories.map((c) => ({ label: c.name, value: c.id }))}
            />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="选填" />
          </Form.Item>
          <Form.Item name="image" label="图片 URL">
            <Input placeholder="选填" />
          </Form.Item>
          <Form.Item label="规格（如辣度、份量加价）">
            <Form.List name="specs">
              {(fields, { add, remove }) => (
                <>
                  {fields.map(({ key, name, ...rest }) => (
                    <Space key={key} style={{ marginBottom: 8 }} align="baseline">
                      <Form.Item {...rest} name={[name, 'spec_type']} rules={[{ required: true }]} noStyle>
                        <Input placeholder="类型（如辣度）" style={{ width: 100 }} />
                      </Form.Item>
                      <Form.Item {...rest} name={[name, 'spec_value']} rules={[{ required: true }]} noStyle>
                        <Input placeholder="选项（如微辣）" style={{ width: 100 }} />
                      </Form.Item>
                      <Form.Item {...rest} name={[name, 'price_delta']} noStyle>
                        <Input type="number" step={0.01} placeholder="加价" style={{ width: 80 }} />
                      </Form.Item>
                      <Button type="text" danger onClick={() => remove(name)}>
                        删除
                      </Button>
                    </Space>
                  ))}
                  <Button type="dashed" onClick={() => add()} block>
                    + 添加规格
                  </Button>
                </>
              )}
            </Form.List>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

// ============ 页面入口 ============

export default function MenuManage() {
  return (
    <div>
      <Typography.Title level={4}>菜单管理</Typography.Title>
      <Tabs
        defaultActiveKey="category"
        items={[
          { key: 'category', label: '菜单分类', children: <CategoryTab /> },
          { key: 'menu', label: '菜品列表', children: <MenuListTab /> },
        ]}
      />
    </div>
  )
}
