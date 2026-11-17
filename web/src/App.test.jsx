import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor, within } from '@testing-library/react';
import React from 'react';
import userEvent from '@testing-library/user-event';
import App from './App.jsx';

const menuItems = [
  { id: 1, name: 'Chicken Karahi', price_pkr: 1450, category: 'Karahi' },
  { id: 2, name: 'Plain Naan', price_pkr: 70, category: 'Bread' },
];

const orders = [
  {
    id: 5,
    status: 'new',
    total_pkr: 1520,
    created_at: '2025-01-01T12:00:00Z',
    items: [{ menu_item_id: 1, qty: 1, name: 'Chicken Karahi' }],
  },
  {
    id: 4,
    status: 'cooking',
    total_pkr: 70,
    created_at: '2025-01-01T12:01:00Z',
    items: [{ menu_item_id: 2, qty: 1, name: 'Plain Naan' }],
  },
  {
    id: 3,
    status: 'ready',
    total_pkr: 70,
    created_at: '2025-01-01T12:02:00Z',
    items: [{ menu_item_id: 2, qty: 1, name: 'Plain Naan' }],
  },
];

function mockFetch() {
  return vi.spyOn(globalThis, 'fetch').mockImplementation(async (url, options) => {
    if (url === '/api/menu') {
      return Response.json({ source: 'db', items: menuItems });
    }
    if (url === '/api/orders' && options?.method === 'POST') {
      return Response.json({ id: 12, status: 'new' }, { status: 201 });
    }
    if (url === '/api/orders') {
      return Response.json(orders);
    }
    if (url === '/api/stats') {
      return Response.json({
        orders_per_minute: 3,
        top_dishes: [
          { name: 'Chicken Karahi', count: 8 },
          { name: 'Plain Naan', count: 4 },
        ],
      });
    }
    throw new Error(`Unexpected API URL: ${url}`);
  });
}

describe('Dhaba Orders', () => {
  afterEach(() => {
    cleanup();
  });

  beforeEach(() => {
    mockFetch();
  });

  it('renders the menu returned by the relative menu API', async () => {
    render(<App />);

    expect(await screen.findByRole('heading', { name: 'Chicken Karahi' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Plain Naan' })).toBeInTheDocument();
    expect(fetch).toHaveBeenCalledWith('/api/menu');
  });

  it('calculates the cart total as quantities change', async () => {
    const user = userEvent.setup();
    render(<App />);

    await user.click(await screen.findByRole('button', { name: 'Add Chicken Karahi' }));
    await user.click(screen.getByRole('button', { name: 'Add Plain Naan' }));

    expect(screen.getByText('Rs 1,520')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Add one Chicken Karahi' }));
    expect(screen.getByText('Rs 2,970')).toBeInTheDocument();
  });

  it('places an order with menu item IDs and quantities', async () => {
    const user = userEvent.setup();
    render(<App />);

    await user.click(await screen.findByRole('button', { name: 'Add Plain Naan' }));
    await user.click(screen.getByRole('button', { name: 'Place order' }));

    await waitFor(() => expect(fetch).toHaveBeenCalledWith('/api/orders', expect.objectContaining({
      method: 'POST',
      body: JSON.stringify({ items: [{ menu_item_id: 2, qty: 1 }] }),
    })));
    expect(await screen.findByText('Order #12 is in the kitchen.')).toBeInTheDocument();
  });

  it('shows kitchen orders in new, cooking, and ready columns', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.click(screen.getByRole('button', { name: 'Kitchen' }));

    const newColumn = await screen.findByRole('region', { name: 'new orders' });
    const cookingColumn = screen.getByRole('region', { name: 'cooking orders' });
    const readyColumn = screen.getByRole('region', { name: 'ready orders' });
    expect(within(newColumn).getByText('Order #5')).toBeInTheDocument();
    expect(within(cookingColumn).getByText('Order #4')).toBeInTheDocument();
    expect(within(readyColumn).getByText('Order #3')).toBeInTheDocument();
  });

  it('renders the stats view with order rate and popular dishes', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.click(screen.getByRole('button', { name: 'Stats' }));

    expect(await screen.findByText('Orders this minute')).toBeInTheDocument();
    expect(screen.getByText('Chicken Karahi')).toBeInTheDocument();
    expect(screen.getByText('8 ordered')).toBeInTheDocument();
    expect(fetch).toHaveBeenCalledWith('/api/stats');
  });
});
