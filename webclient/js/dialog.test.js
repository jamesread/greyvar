import { describe, it, expect, vi } from 'vitest'

vi.mock('./menu/DialogPreview.js', () => ({ createDialogPreview: () => null }))
const { dialogMenu, offersWear } = await import('./dialog.js')

describe('dialogMenu', () => {
  const dialog = {
    entityId: 42,
    node: 'hello',
    speaker: 'Bill',
    text: 'Morning!',
    options: [{ label: 'Got any seeds?' }, { label: 'Bye', ends: true }]
  }

  it('shows the speaker, the line and one item per option', () => {
    const menu = dialogMenu(dialog, () => {}, () => {})
    expect(menu.title).toBe('Bill')
    expect(menu.subtitle).toBe('Morning!')
    expect(menu.items.map((i) => i.label)).toEqual(['Got any seeds?', 'Bye'])
  })

  it('answers with the option index; ending options close the menu', () => {
    const sent = []
    let closed = 0
    const menu = dialogMenu(dialog, (...args) => sent.push(args), () => { closed++ })
    menu.items[0].action()
    expect(sent).toEqual([[42, 'hello', 0]])
    expect(closed).toBe(0)
    menu.items[1].action()
    expect(sent[1]).toEqual([42, 'hello', 1])
    expect(closed).toBe(1)
  })
})

describe('customization dialogs', () => {
  const tailor = {
    entityId: 7,
    node: 'shirts',
    speaker: 'Geraldine',
    options: [{ label: 'Blue', wear: { shirt: 'blue' } }, { label: 'Never mind' }]
  }

  it('only steps that offer look changes get a live preview', () => {
    expect(offersWear(tailor)).toBe(true)
    expect(offersWear({ options: [{ label: 'Bye', ends: true }] })).toBe(false)
    const seen = []
    const preview = (_ctx, wearFor) => { seen.push(wearFor('option-0'), wearFor('option-1')); return {} }
    expect(dialogMenu(tailor, () => {}, () => {}, preview).decorate).toBeTypeOf('function')
    dialogMenu(tailor, () => {}, () => {}, preview).decorate({})
    expect(seen).toEqual([{ shirt: 'blue' }, null])
    expect(dialogMenu({ options: [{ label: 'Bye' }] }, () => {}, () => {}, preview).decorate).toBe(null)
  })
})
