import { useEffect, useState } from 'react'
import { Tabs, Card, Button, Radio, Form, message, Empty } from 'antd'
import { PlusOutlined, MinusOutlined } from '@ant-design/icons'
import type { Menu, MenuCategory, MenuSpec, Table as TableType } from '../api/types'
import { listMenuCategories, listMenus } from '../api/menu'
import { listTables } from '../api/table'
import { createOrder } from '../api/order'

interface CartItem {
  menuId: number
  name: string
  price: number
  quantity: number
  specInfo?: string
}

/** 按 spec_type 分组，每组取第一个作为默认 */
function defaultSpecs(specs: MenuSpec[]): MenuSpec[] {
  const byType = new Map<string, MenuSpec>()
  for (const s of specs) {
    if (!byType.has(s.spec_type)) byType.set(s.spec_type, s)
  }
  return Array.from(byType.values())
}

const ORDER_TYPE_OPTIONS = [
  { label: '堂食', value: 'dine_in' },
  { label: '打包外带', value: 'takeaway' },
]

export default function OrderDesk() {
  const [categories, setCategories] = useState<MenuCategory[]>([])
  const [menus, setMenus] = useState<Menu[]>([])
  const [loading, setLoading] = useState(false)
  const [activeCategory, setActiveCategory] = useState<string>('all')
  const [cart, setCart] = useState<CartItem[]>([])
  const [tables, setTables] = useState<TableType[]>([])
  const [tablesLoading, setTablesLoading] = useState(false)
  const [orderType, setOrderType] = useState<string>('dine_in')
  const [tableId, setTableId] = useState<number | undefined>(undefined)
  const [submitting, setSubmitting] = useState(false)
  /** 每个菜品卡片上的数量与选中规格：menuId -> { quantity, selectedSpecs } */
  const [cardSelections, setCardSelections] = useState<Record<number, { quantity: number; selectedSpecs: MenuSpec[] }>>({})

  const loadCategories = async () => {
    try {
      const res = await listMenuCategories({ current: 1, pageSize: 100 })
      setCategories(Array.isArray(res?.categories) ? res.categories : [])
    } catch {
      message.error('加载分类失败')
      setCategories([])
    }
  }

  const loadMenus = async () => {
    setLoading(true)
    try {
      const res = await listMenus({
        current: 1,
        pageSize: 500,
        category: activeCategory === 'all' ? undefined : activeCategory,
      })
      setMenus(Array.isArray(res?.menus) ? res.menus : [])
    } catch {
      message.error('加载菜品失败')
      setMenus([])
    } finally {
      setLoading(false)
    }
  }

  const loadTables = async () => {
    setTablesLoading(true)
    try {
      const res = await listTables({ current: 1, pageSize: 500 })
      setTables(Array.isArray(res?.tables) ? res.tables : [])
    } catch {
      message.error('加载桌号列表失败')
      setTables([])
    } finally {
      setTablesLoading(false)
    }
  }

  useEffect(() => {
    loadCategories()
    loadTables() /* 进入点餐页即拉取桌号，默认堂食需要 */
  }, [])

  useEffect(() => {
    loadMenus()
  }, [activeCategory])

  /** 选堂食时拉取桌号列表，确保有选项 */
  useEffect(() => {
    if (orderType === 'dine_in') {
      loadTables()
    }
  }, [orderType])

  const getSelection = (menu: Menu) => {
    const raw = cardSelections[menu.id]
    const defaults = menu.specs?.length ? defaultSpecs(menu.specs) : []
    const byType = new Map(defaults.map((s) => [s.spec_type, s]))
    raw?.selectedSpecs?.forEach((s) => byType.set(s.spec_type, s))
    const selectedSpecs = Array.from(byType.values())
    const quantity = raw?.quantity ?? 1
    return { quantity, selectedSpecs }
  }

  const setMenuQuantity = (menu: Menu, quantity: number) => {
    setCardSelections((prev) => ({
      ...prev,
      [menu.id]: {
        quantity: Math.max(1, quantity),
        selectedSpecs: prev[menu.id]?.selectedSpecs ?? (menu.specs?.length ? defaultSpecs(menu.specs) : []),
      },
    }))
  }

  const setMenuSpec = (menu: Menu, specType: string, spec: MenuSpec) => {
    const current = getSelection(menu).selectedSpecs.filter((s) => s.spec_type !== specType)
    setCardSelections((prev) => ({
      ...prev,
      [menu.id]: {
        quantity: prev[menu.id]?.quantity ?? 1,
        selectedSpecs: [...current, spec],
      },
    }))
  }

  const addToCart = (menu: Menu) => {
    const { quantity, selectedSpecs } = getSelection(menu)
    const specInfo = selectedSpecs.length ? selectedSpecs.map((s) => `${s.spec_type}:${s.spec_value}`).join(' ') : undefined
    const unitPrice = menu.price + selectedSpecs.reduce((sum, s) => sum + (s.price_delta ?? 0), 0)
    const price = Math.round(unitPrice * 100) / 100
    setCart((prev) => {
      const exist = prev.find((i) => i.menuId === menu.id && (i.specInfo ?? '') === (specInfo ?? ''))
      if (exist) {
        return prev.map((i) =>
          i.menuId === menu.id && (i.specInfo ?? '') === (specInfo ?? '') ? { ...i, quantity: i.quantity + quantity } : i
        )
      }
      return [...prev, { menuId: menu.id, name: menu.name, price, quantity, specInfo }]
    })
  }

  const updateCartQty = (menuId: number, specInfo: string | undefined, delta: number) => {
    setCart((prev) =>
      prev
        .map((i) =>
          i.menuId === menuId && (i.specInfo ?? '') === (specInfo ?? '') ? { ...i, quantity: i.quantity + delta } : i
        )
        .filter((i) => i.quantity > 0)
    )
  }

  const cartTotal = cart.reduce((sum, i) => sum + i.price * i.quantity, 0)

  const handleSubmit = async () => {
    if (cart.length === 0) {
      message.warning('请先添加菜品')
      return
    }
    setSubmitting(true)
    try {
      await createOrder({
        order_type: orderType,
        table_id: orderType === 'dine_in' ? tableId : undefined,
        items: cart.map((i) => ({
          menu_name: i.name,
          quantity: i.quantity,
          unit_price: i.price,
          spec_info: i.specInfo || undefined,
        })),
        total_amount: Math.round(cartTotal * 100) / 100,
      })
      message.success('下单成功')
      setCart([])
    } catch {
      message.error('下单失败')
    } finally {
      setSubmitting(false)
    }
  }

  const tabItems = [
    { key: 'all', label: '全部' },
    ...categories.map((c) => ({ key: c.name, label: c.name })),
  ]

  return (
    <div style={{ display: 'flex', gap: 16, height: 'calc(100vh - 120px)', minHeight: 400 }}>
      {/* 左侧：购物车固定展开 */}
      <div
        style={{
          width: 320,
          flexShrink: 0,
          display: 'flex',
          flexDirection: 'column',
          background: '#fafafa',
          borderRadius: 8,
          border: '1px solid #f0f0f0',
          overflow: 'hidden',
        }}
      >
        <div style={{ padding: '12px 16px', borderBottom: '1px solid #f0f0f0', fontWeight: 600 }}>
          购物车
        </div>
        <div style={{ flex: 1, overflow: 'auto', padding: 12 }}>
          {cart.length === 0 ? (
            <Empty description="暂无菜品" image={Empty.PRESENTED_IMAGE_SIMPLE} />
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {cart.map((item, idx) => (
                <div
                  key={`${item.menuId}-${item.specInfo ?? ''}-${idx}`}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: 10,
                    background: '#fff',
                    borderRadius: 8,
                    boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
                  }}
                >
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: 500 }}>{item.name}</div>
                    {item.specInfo ? (
                      <div style={{ fontSize: 11, color: '#888', marginTop: 2 }}>{item.specInfo}</div>
                    ) : null}
                    <div style={{ color: '#f5222d', fontSize: 13, marginTop: 2 }}>¥{item.price.toFixed(2)}</div>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <Button
                      type="text"
                      size="small"
                      icon={<MinusOutlined />}
                      onClick={() => updateCartQty(item.menuId, item.specInfo, -1)}
                    />
                    <span style={{ minWidth: 28, textAlign: 'center', fontWeight: 500 }}>{item.quantity}</span>
                    <Button
                      type="text"
                      size="small"
                      icon={<PlusOutlined />}
                      onClick={() => updateCartQty(item.menuId, item.specInfo, 1)}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
        <div style={{ padding: 16, borderTop: '1px solid #f0f0f0' }}>
          <div style={{ fontSize: 16, fontWeight: 600, marginBottom: 12 }}>
            合计：¥{cartTotal.toFixed(2)}
          </div>
          <Form layout="vertical" style={{ marginBottom: 12 }}>
            <Form.Item label="就餐方式" style={{ marginBottom: 8 }}>
              <Radio.Group
                value={orderType}
                onChange={(e) => setOrderType(e.target.value)}
                optionType="button"
                options={ORDER_TYPE_OPTIONS}
                style={{ width: '100%' }}
              />
            </Form.Item>
            {orderType === 'dine_in' && (
              <Form.Item label="桌号" style={{ marginBottom: 8 }}>
                {tablesLoading ? (
                  <span style={{ color: '#888', fontSize: 12 }}>加载桌号中…</span>
                ) : tables.length === 0 ? (
                  <span style={{ color: '#888', fontSize: 12 }}>暂无餐桌，请先在「餐桌管理」添加</span>
                ) : (
                  <Radio.Group
                    value={tableId}
                    onChange={(e) => setTableId(e.target.value)}
                    optionType="button"
                    style={{ width: '100%', display: 'flex', flexWrap: 'wrap', gap: 8 }}
                    options={tables.map((t) => ({ label: `${t.code}（${t.capacity}人）`, value: t.id }))}
                  />
                )}
              </Form.Item>
            )}
          </Form>
          <Button
            type="primary"
            block
            size="large"
            onClick={handleSubmit}
            loading={submitting}
            disabled={cart.length === 0}
          >
            提交订单
          </Button>
        </div>
      </div>

      {/* 右侧：分类 + 菜品 */}
      <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        <Tabs
          activeKey={activeCategory}
          onChange={setActiveCategory}
          items={tabItems}
          style={{ marginBottom: 12, flexShrink: 0 }}
        />
        <div style={{ flex: 1, overflow: 'auto' }}>
          {menus.length === 0 && !loading ? (
            <Empty description="暂无菜品" />
          ) : (
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))',
                gap: 16,
              }}
            >
              {menus.map((menu) => {
                const { quantity, selectedSpecs } = getSelection(menu)
                const unitPrice = menu.price + selectedSpecs.reduce((sum, s) => sum + (s.price_delta ?? 0), 0)
                const specsByType = new Map<string, MenuSpec[]>()
                for (const s of menu.specs ?? []) {
                  if (!specsByType.has(s.spec_type)) specsByType.set(s.spec_type, [])
                  specsByType.get(s.spec_type)!.push(s)
                }
                return (
                  <Card
                    key={menu.id}
                    hoverable
                    style={{ borderRadius: 12, overflow: 'hidden' }}
                    styles={{ body: { padding: 12 } }}
                    cover={
                      menu.image ? (
                        <img
                          alt={menu.name}
                          src={menu.image}
                          style={{ height: 130, objectFit: 'cover' }}
                          onError={(e) => {
                            (e.target as HTMLImageElement).style.display = 'none'
                          }}
                        />
                      ) : (
                        <div
                          style={{
                            height: 130,
                            background: 'linear-gradient(135deg, #f5f5f5 0%, #e8e8e8 100%)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            color: '#999',
                            fontSize: 12,
                          }}
                        >
                          暂无图片
                        </div>
                      )
                    }
                  >
                    <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', gap: 8, marginBottom: 10 }}>
                      <span style={{ fontWeight: 600, fontSize: 15, flex: 1, minWidth: 0 }}>{menu.name}</span>
                      <span style={{ color: '#f5222d', fontSize: 16, fontWeight: 600, flexShrink: 0 }}>
                        ¥{unitPrice.toFixed(2)}
                        {selectedSpecs.length > 0 && (
                          <span style={{ fontSize: 11, color: '#888', fontWeight: 400 }}> 起</span>
                        )}
                      </span>
                    </div>
                    {/* 数量：+- */}
                    <div style={{ marginBottom: 10, display: 'flex', alignItems: 'center', gap: 8 }}>
                      <span style={{ fontSize: 12, color: '#666', width: 36 }}>数量</span>
                      <Button
                        type="text"
                        size="small"
                        icon={<MinusOutlined />}
                        onClick={() => setMenuQuantity(menu, Math.max(1, quantity - 1))}
                        disabled={quantity <= 1}
                      />
                      <span style={{ minWidth: 28, textAlign: 'center', fontWeight: 600 }}>{quantity}</span>
                      <Button
                        type="text"
                        size="small"
                        icon={<PlusOutlined />}
                        onClick={() => setMenuQuantity(menu, Math.min(99, quantity + 1))}
                        disabled={quantity >= 99}
                      />
                    </div>
                    {/* 规格：规格名+规格值同一行 */}
                    {specsByType.size > 0 &&
                      Array.from(specsByType.entries()).map(([type, opts]) => {
                        const current = selectedSpecs.find((s) => s.spec_type === type)
                        return (
                          <div key={type} style={{ marginBottom: 8, display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                            <span style={{ fontSize: 12, color: '#666', flexShrink: 0 }}>{type}</span>
                            <Radio.Group
                              value={current?.spec_value}
                              onChange={(e) => {
                                const spec = opts.find((o) => o.spec_value === e.target.value)
                                if (spec) setMenuSpec(menu, type, spec)
                              }}
                              optionType="button"
                              size="small"
                              options={opts.map((o) => ({
                                label: `${o.spec_value}${o.price_delta ? ` +¥${o.price_delta}` : ''}`,
                                value: o.spec_value,
                              }))}
                            />
                          </div>
                        )
                      })}
                    <Button
                      type="primary"
                      block
                      size="middle"
                      icon={<PlusOutlined />}
                      onClick={() => addToCart(menu)}
                      style={{ marginTop: 8, borderRadius: 8 }}
                    >
                      加入购物车
                    </Button>
                  </Card>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
