// NPC conversations: the server sends a Dialog (speaker, text, options); it
// shows as a menu, and picking an option answers with a DialogChoiceRequest.
// The server sends the next step, or the option ends the conversation.

import { createDialogPreview } from './menu/DialogPreview.js'

export const DIALOG_MENU_ID = 'dialog'

/** Whether a Dialog step offers look changes (tailor, skin tone, hair). */
export function offersWear (dialog) {
  return (dialog?.options ?? []).some((o) => Object.keys(o.wear ?? {}).length > 0)
}

/**
 * Menu config for a server Dialog message. With createPreview, steps that
 * offer look changes show a live preview of the focused option.
 */
export function dialogMenu (dialog, send, close, createPreview = null) {
  const wearFor = (itemId) => {
    const index = Number(String(itemId ?? '').replace('option-', ''))
    return dialog.options?.[index]?.wear ?? null
  }
  return {
    id: DIALOG_MENU_ID,
    title: dialog.speaker ?? '',
    subtitle: dialog.text ?? '',
    decorate: createPreview != null && offersWear(dialog)
      ? (ctx) => createPreview(ctx, wearFor)
      : null,
    items: (dialog.options ?? []).map((option, index) => ({
      id: `option-${index}`,
      label: option.label ?? '',
      action: () => {
        send(dialog.entityId, dialog.node, index)
        if (option.ends) {
          close()
        }
      }
    }))
  }
}

export function showDialog (dialog) {
  const ms = window.menuSystem
  if (ms == null || dialog == null) {
    return
  }
  ms.registerMenu(dialogMenu(
    dialog,
    (entityId, node, option) => window.serverConnection?.sendDialogChoice(entityId, node, option),
    () => ms.close(),
    createDialogPreview
  ))
  ms.open(DIALOG_MENU_ID)
}
