/** 与后端 happyeat API 对齐的通用类型（基于 happyeat.json） */

export interface LoginReq {
  username: string
  password: string
}

export interface LoginReply {
  access_token: string
  expire: number
}

/** 菜单分类（与后端 MenuCategory 一致） */
export interface MenuCategory {
  id: number
  name: string
  description?: string
}

/** 兼容旧名 */
export type Category = MenuCategory

export interface MenuSpec {
  spec_type: string
  spec_value: string
  price_delta: number
}

export interface Menu {
  id: number
  name: string
  price: number
  category_id: number
  description?: string
  image?: string
  specs?: MenuSpec[]
  create_at: number
  update_at: number
}

/** 餐桌分类 */
export interface TableCategory {
  id: number
  name: string
  description?: string
}

export interface Table {
  id: number
  code: string
  status: string
  capacity: number
  category_id: number
  qr_code?: string
  create_at: number
  update_at: number
}

export interface OrderItem {
  menu_name: string
  quantity: number
  unit_price: number
  amount: number
  spec_info?: string
}

export interface Order {
  id: number
  order_no: string
  order_type: string
  status: string
  total_amount: number
  table_id?: number
  remark?: string
  items?: OrderItem[]
  create_at: number
  update_at: number
}
