import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { server } from '../../tests/mocks/server'
import { CalculatorPage } from './CalculatorPage'

/** Integration tests: View → useCalculator → calculatorService → (mocked) HTTP. */
function renderPage() {
  const user = userEvent.setup()
  render(<CalculatorPage />)
  const keypad = screen.getByRole('group', { name: 'Keypad' })
  const press = async (...names: string[]) => {
    for (const name of names) await user.click(within(keypad).getByRole('button', { name }))
  }
  return { user, press }
}

const display = () => screen.getByRole('status', { name: 'Result' })

describe('CalculatorPage', () => {
  it('shows the API status', async () => {
    renderPage()
    expect(await screen.findByText('API online')).toBeInTheDocument()
  })

  it('shows when the API is offline', async () => {
    server.use(http.get('/api/health', () => HttpResponse.error()))
    renderPage()
    expect(await screen.findByText('API offline')).toBeInTheDocument()
  })

  it('performs a calculation with the keypad', async () => {
    const { press } = renderPage()
    await press('1', '2', 'Multiply', '3', 'Equals')

    expect(await screen.findByText('12 × 3 =')).toBeInTheDocument()
    expect(display()).toHaveTextContent('36')
    expect(screen.getByRole('listitem')).toHaveTextContent('12 × 3= 36')
  })

  it('shows the division by zero error from the API', async () => {
    const { press } = renderPage()
    await press('9', 'Divide', '0', 'Equals')

    expect(await screen.findByRole('alert')).toHaveTextContent('Division by zero')
    expect(screen.getByText('400')).toBeInTheDocument()
  })

  it('shows a validation error when the second number is missing', async () => {
    const { press } = renderPage()
    await press('9', 'Add', 'Equals')

    expect(screen.getByRole('alert')).toHaveTextContent('Enter a second number')
  })

  it('shows a friendly message when the API is unreachable', async () => {
    server.use(http.post('/api/calculate', () => HttpResponse.error()))
    const { press } = renderPage()
    await press('1', 'Add', '1', 'Equals')

    expect(await screen.findByRole('alert')).toHaveTextContent('Cannot reach the calculator service')
    expect(screen.getByText('network error')).toBeInTheDocument()
  })

  it('supports the physical keyboard', async () => {
    const { user } = renderPage()
    await user.keyboard('7-2{Enter}')

    expect(await screen.findByText('7 − 2 =')).toBeInTheDocument()
    expect(display()).toHaveTextContent('5')

    await user.keyboard('{Escape}')
    expect(display()).toHaveTextContent('0')
  })

  it('reveals advanced operations', async () => {
    const { user, press } = renderPage()
    const toggle = screen.getByRole('button', { name: 'Advanced' })
    expect(screen.queryByRole('group', { name: 'Advanced operations' })).not.toBeInTheDocument()

    await user.click(toggle)
    expect(toggle).toHaveAttribute('aria-pressed', 'true')

    await press('1', '6')
    await user.click(screen.getByRole('button', { name: 'Square root' }))
    expect(await screen.findByText('√16 =')).toBeInTheDocument()
    expect(display()).toHaveTextContent('4')
  })

  it('clears the history', async () => {
    const { user, press } = renderPage()
    await press('1', 'Add', '1', 'Equals')
    expect(await screen.findByRole('listitem')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Clear' }))
    expect(screen.getByText('No calculations yet')).toBeInTheDocument()
  })
})
