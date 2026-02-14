/**
 * 菜单相关 API
 * 这里集中所有「菜单分类」和「菜品」的请求，和后端的 /central/v1/menu/*、/menus 对应。
 * 用 axios 的 api 实例（已在 client.ts 里挂好 baseURL 和 token），所以这里只写路径和参数。
 */
import { api } from './client'
import type { Menu, MenuCategory, MenuSpec } from './types'

// ============ 菜单分类 ============

export interface ListCategoriesParams {
  current?: number
  pageSize?: number
  name?: string
}

export interface ListCategoriesRes {
  categories: MenuCategory[]
  total: number
}

/** 拉取分类列表（GET，参数走 query） */
export async function listMenuCategories(params?: ListCategoriesParams): Promise<ListCategoriesRes> {
  const { data } = await api.get<ListCategoriesRes>('/central/v1/menu/categories', { params })
  return data
}

/** 创建分类：后端只收 name、description */
export async function createMenuCategory(body: { name: string; description?: string }): Promise<void> {
  await api.post('/central/v1/menu/category', body)
}

/** 获取单个分类 */
export async function getMenuCategory(id: number): Promise<{ category: MenuCategory }> {
  const { data } = await api.get<{ category: MenuCategory }>(`/central/v1/menu/category/${id}`)
  return data
}

/** 更新分类：后端要整份 category（含 id） */
export async function updateMenuCategory(id: number, category: MenuCategory): Promise<void> {
  await api.put(`/central/v1/menu/category/${id}`, { category })
}

/** 删除分类 */
export async function deleteMenuCategory(id: number): Promise<void> {
  await api.delete(`/central/v1/menu/category/${id}`)
}

// ============ 菜品 ============

export interface ListMenusParams {
  current?: number
  pageSize?: number
  name?: string
  category?: string
}

export interface ListMenusRes {
  menus: Menu[]
  total: number
}

/** 拉取菜品列表 */
export async function listMenus(params?: ListMenusParams): Promise<ListMenusRes> {
  const { data } = await api.get<ListMenusRes>('/central/v1/menus', { params })
  return data
}

export interface CreateMenuBody {
  name: string
  price: number
  category_id: number
  description?: string
  image?: string
  specs?: MenuSpec[]
}

/** 创建菜品 */
export async function createMenu(body: CreateMenuBody): Promise<void> {
  await api.post('/central/v1/menus', body)
}

/** 获取单个菜品 */
export async function getMenu(id: number): Promise<{ menu: Menu }> {
  const { data } = await api.get<{ menu: Menu }>(`/central/v1/menu/${id}`)
  return data
}

/** 更新菜品：后端要整份 menu（含 id） */
export async function updateMenu(id: number, menu: Menu): Promise<void> {
  await api.put(`/central/v1/menu/${id}`, { menu })
}

/** 删除菜品 */
export async function deleteMenu(id: number): Promise<void> {
  await api.delete(`/central/v1/menu/${id}`)
}
