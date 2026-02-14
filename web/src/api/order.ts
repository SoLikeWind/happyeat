import { api } from './client'
import type { Order } from './types'

export interface CreateOrderItem {
  menu_name: string
  quantity: number
  unit_price: number
  spec_info?: string
}

export async function listOrders(params?: {
  current?: number
  pageSize?: number
  status?: string
  order_type?: string
  table_id?: number
}) {
  const { data } = await api.get<{ orders: Order[]; total: number }>('/central/v1/orders', { params })
  return data
}

export async function getOrder(id: number) {
  const { data } = await api.get<{ order: Order }>(`/central/v1/order/${id}`)
  return data
}

export async function createOrder(body: {
  order_type: string
  table_id?: number
  items: CreateOrderItem[]
  total_amount: number
  remark?: string
}) {
  const { data } = await api.post<{ order: Order }>('/central/v1/orders', body)
  return data
}

export async function updateOrderStatus(id: number, status: string) {
  await api.put(`/central/v1/order/${id}/status`, { status })
}

export async function listWorkbenchOrders(params?: { current?: number; pageSize?: number; status?: string }) {
  const { data } = await api.get<{ orders: Order[]; total: number }>('/central/v1/workbench/orders', { params })
  return data
}
