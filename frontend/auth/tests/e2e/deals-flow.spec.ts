import { expect, test } from '@playwright/test'

async function mockCommonApi(page: import('@playwright/test').Page) {
  let createdDeal: { id: string; customer_id: string; vehicle_id: string; amount: string; stage: string } | null = null
  await page.route('**/api/**', async (route) => {
    const req = route.request()
    const url = new URL(req.url())
    const path = url.pathname
    const method = req.method()

    if (path === '/api/login' && method === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          user_id: 'user-1',
          email: 'manager@test.local',
          access_token: 'access-token',
          refresh_token: 'refresh-token',
          expires_at: Date.now() + 3600_000,
        }),
      })
      return
    }

    if (path === '/api/logout' && method === 'POST') {
      await route.fulfill({ status: 204, body: '' })
      return
    }

    if (path === '/api/customers' && method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          customers: [{ id: 'c-1', name: 'Иван Петров', email: 'ivan@example.com', phone: '', customer_type: 'individual', inn: '', address: '', notes: '', created_at: 1, updated_at: 1 }],
          total: 1,
        }),
      })
      return
    }

    if (path === '/api/vehicles' && method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          vehicles: [{ id: 'v-1', vin: 'VIN123456789', make: 'Toyota', model: 'Camry', year: 2022, mileage_km: 0, price: '2500000', status: 'available', color: 'black', notes: '', created_at: 1, updated_at: 1 }],
          total: 1,
        }),
      })
      return
    }

    if (path === '/api/deals' && method === 'POST') {
      const payload = req.postDataJSON() as { customer_id: string; vehicle_id: string; amount?: string; stage?: string }
      createdDeal = {
        id: 'd-1',
        customer_id: payload.customer_id,
        vehicle_id: payload.vehicle_id,
        amount: payload.amount || '0',
        stage: payload.stage || 'draft',
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ ...createdDeal, assigned_to: '', notes: '', created_at: 1, updated_at: 1 }),
      })
      return
    }

    if (path === '/api/deals' && method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          deals: createdDeal
            ? [{ ...createdDeal, assigned_to: '', notes: '', created_at: 1, updated_at: 1 }]
            : [],
          total: createdDeal ? 1 : 0,
        }),
      })
      return
    }

    if (path === '/api/me' && method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ user_id: 'user-1', email: 'manager@test.local', valid: true }),
      })
      return
    }

    await route.fulfill({ status: 404, body: 'not mocked' })
  })
}

test('smoke: login -> create deal', async ({ page }) => {
  await mockCommonApi(page)
  await page.goto('/login')
  await page.getByLabel('Email').fill('manager@test.local')
  await page.getByLabel('Пароль').fill('secret')
  await page.getByRole('button', { name: 'Войти' }).click()
  await expect(page).toHaveURL(/\/customers$/)

  await page.getByRole('link', { name: 'Сделки' }).click()
  await page.getByRole('link', { name: '+ Новая сделка' }).click()
  await expect(page.getByRole('heading', { name: 'Новая сделка' })).toBeVisible()

  await page.getByLabel('Клиент *').selectOption('c-1')
  await page.getByLabel('Автомобиль *').selectOption('v-1')
  await page.getByLabel('Сумма').fill('1250000')
  await page.getByRole('button', { name: 'Создать' }).click()
  await expect(page).toHaveURL(/\/deals$/)
  await expect(page.getByText('1 250 000')).toBeVisible()
})

test('negative: expired token on deals triggers redirect to login', async ({ page }) => {
  await page.addInitScript(() => {
    sessionStorage.setItem('dealer_access_token', 'expired')
    sessionStorage.setItem('dealer_refresh_token', 'expired-r')
  })
  await page.route('**/api/**', async (route) => {
    const req = route.request()
    const path = new URL(req.url()).pathname
    if (path === '/api/me') {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user_id: 'u-1', email: 'u@test.local', valid: true }) })
      return
    }
    if (path === '/api/deals') {
      await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ error: 'unauthorized' }) })
      return
    }
    if (path === '/api/logout') {
      await route.fulfill({ status: 204, body: '' })
      return
    }
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({}) })
  })
  await page.goto('/deals')
  await expect(page).toHaveURL(/\/login$/)
})

test('negative: insufficient permissions on deals triggers redirect', async ({ page }) => {
  await page.addInitScript(() => {
    sessionStorage.setItem('dealer_access_token', 'valid')
    sessionStorage.setItem('dealer_refresh_token', 'valid-r')
  })
  await page.route('**/api/**', async (route) => {
    const req = route.request()
    const path = new URL(req.url()).pathname
    if (path === '/api/me') {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user_id: 'u-2', email: 'u2@test.local', valid: true }) })
      return
    }
    if (path === '/api/deals') {
      await route.fulfill({ status: 403, contentType: 'application/json', body: JSON.stringify({ error: 'forbidden' }) })
      return
    }
    if (path === '/api/logout') {
      await route.fulfill({ status: 204, body: '' })
      return
    }
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({}) })
  })
  await page.goto('/deals')
  await expect(page).toHaveURL(/\/login$/)
})

test('negative: api failure keeps ui stable and shows error', async ({ page }) => {
  await page.addInitScript(() => {
    sessionStorage.setItem('dealer_access_token', 'valid')
    sessionStorage.setItem('dealer_refresh_token', 'valid-r')
  })
  await page.route('**/api/**', async (route) => {
    const req = route.request()
    const path = new URL(req.url()).pathname
    if (path === '/api/me') {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user_id: 'u-3', email: 'u3@test.local', valid: true }) })
      return
    }
    if (path === '/api/deals') {
      await route.fulfill({ status: 500, contentType: 'application/json', body: JSON.stringify({ error: 'internal error' }) })
      return
    }
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({}) })
  })
  await page.goto('/deals')
  await expect(page.getByText('internal error')).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Сделки' })).toBeVisible()
})
