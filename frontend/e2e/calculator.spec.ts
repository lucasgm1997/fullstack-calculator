import { expect, test, type Page } from '@playwright/test'

/**
 * Full-stack E2E: real browser → React app → Vite proxy → Go API.
 * Nothing is mocked; results and error messages come from the backend.
 */

function calculator(page: Page) {
  const keypad = page.getByRole('group', { name: 'Keypad' })
  return {
    result: page.getByRole('status', { name: 'Result' }),
    alert: page.getByRole('alert'),
    history: page.getByRole('region', { name: 'History' }).getByRole('listitem'),
    async press(...keys: string[]) {
      for (const name of keys) await keypad.getByRole('button', { name, exact: true }).click()
    },
  }
}

test.beforeEach(async ({ page }) => {
  await page.goto('/')
  // The health check proves the browser reaches the Go API through the proxy.
  await expect(page.getByText('API online')).toBeVisible()
})

test('calculates with the keypad and records the call', async ({ page }) => {
  const calc = calculator(page)

  await calc.press('1', '2', 'Multiply', '3', 'Equals')

  await expect(calc.result).toHaveText('36')
  await expect(page.getByText('12 × 3 =')).toBeVisible()
  await expect(calc.history).toHaveCount(1)
  await expect(calc.history.first()).toContainText('12 × 3')
  await expect(calc.history.first()).toContainText('= 36')

  // The inspector shows the real request and the backend's real response.
  await page.getByText('Last API call').click()
  await expect(page.getByText('"operation": "multiply"')).toBeVisible()
  await expect(page.getByText('200', { exact: true })).toBeVisible()
  await expect(page.getByText('"result": 36')).toBeVisible()
})

test('chains operations from the keyboard', async ({ page }) => {
  const calc = calculator(page)

  // 2 + 3 × 4 is evaluated left to right by the API: (2 + 3) × 4 = 20
  await page.keyboard.type('2+3*4')
  await page.keyboard.press('Enter')

  await expect(calc.result).toHaveText('20')
  await expect(page.getByText('5 × 4 =')).toBeVisible()
  await expect(calc.history).toHaveCount(2)
})

test('shows the division by zero error and recovers', async ({ page }) => {
  const calc = calculator(page)

  await calc.press('9', 'Divide', '0', 'Equals')

  await expect(calc.alert).toHaveText('Division by zero')
  await page.getByText('Last API call').click()
  await expect(page.getByText('400', { exact: true })).toBeVisible()
  await expect(page.getByText('"error": "division by zero"')).toBeVisible()

  // The input is kept, so the user can fix the divisor and retry.
  await calc.press('Backspace', '3', 'Equals')
  await expect(calc.alert).toBeHidden()
  await expect(calc.result).toHaveText('3')
})

test('validates a missing second number without calling the API', async ({ page }) => {
  const calc = calculator(page)
  let requests = 0
  page.on('request', (request) => {
    if (request.url().includes('/api/calculate')) requests++
  })

  await calc.press('5', 'Add', 'Equals')

  await expect(calc.alert).toHaveText('Enter a second number')
  expect(requests).toBe(0)
})

test('runs advanced operations and backend validation', async ({ page }) => {
  const calc = calculator(page)
  await page.getByRole('button', { name: 'Advanced' }).click()
  const advanced = page.getByRole('group', { name: 'Advanced operations' })

  await calc.press('8', '1')
  await advanced.getByRole('button', { name: 'Square root' }).click()
  await expect(calc.result).toHaveText('9')

  await calc.press('All clear', '1', '5')
  await advanced.getByRole('button', { name: 'Percent of' }).click()
  await calc.press('2', '0', '0', 'Equals')
  await expect(calc.result).toHaveText('30')

  // The frontend has no math rules: this message comes from the Go domain.
  await calc.press('All clear', '4', 'Toggle sign')
  await advanced.getByRole('button', { name: 'Square root' }).click()
  await expect(calc.alert).toHaveText('Square root of a negative number')
})
