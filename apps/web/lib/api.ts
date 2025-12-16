import { z } from 'zod'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

// Schemas
export const ProductSchema = z.object({
  id: z.string().uuid(),
  title: z.string(),
  brand: z.string().nullable().optional(),
  model: z.string().nullable().optional(),
  image_url: z.string().nullable().optional(),
  created_at: z.string(),
  updated_at: z.string(),
})

export const OfferSchema = z.object({
  id: z.string().uuid(),
  product_id: z.string().uuid(),
  source: z.string(),
  seller: z.string(),
  price_amount: z.number(),
  currency: z.string(),
  shipping_to_us_amount: z.number(),
  total_to_us_amount: z.number(),
  est_delivery_days_min: z.number().nullable().optional(),
  est_delivery_days_max: z.number().nullable().optional(),
  in_stock: z.boolean(),
  url: z.string().nullable().optional(),
  fetched_at: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
})

export const ProductWithMinPriceSchema = ProductSchema.extend({
  min_price_cents: z.number().nullable().optional(),
})

export type Product = z.infer<typeof ProductSchema>
export type Offer = z.infer<typeof OfferSchema>
export type ProductWithMinPrice = z.infer<typeof ProductWithMinPriceSchema>

// API functions
export async function searchProducts(query: string): Promise<ProductWithMinPrice[]> {
  const res = await fetch(`${API_URL}/api/search?query=${encodeURIComponent(query)}`)
  if (!res.ok) {
    throw new Error('Failed to search products')
  }
  const data = await res.json()
  return z.object({ products: z.array(ProductWithMinPriceSchema) }).parse(data).products
}

export async function getProduct(id: string): Promise<Product> {
  const res = await fetch(`${API_URL}/api/products/${id}`)
  if (!res.ok) {
    throw new Error('Failed to get product')
  }
  return ProductSchema.parse(await res.json())
}

export async function getProductOffers(id: string): Promise<Offer[]> {
  const res = await fetch(`${API_URL}/api/products/${id}/offers`)
  if (!res.ok) {
    throw new Error('Failed to get offers')
  }
  const data = await res.json()
  return z.object({ offers: z.array(OfferSchema) }).parse(data).offers
}

export async function fetchPrices(source: string): Promise<{ job_id: string; status: string; source: string }> {
  const res = await fetch(`${API_URL}/api/admin/jobs/fetch_prices`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ source }),
  })
  if (!res.ok) {
    throw new Error('Failed to fetch prices')
  }
  return res.json()
}

