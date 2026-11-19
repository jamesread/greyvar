// Reconnect policy and the "disconnected" banner for ServerConnection.

export const RECONNECT_BASE_MS = 500
export const RECONNECT_MAX_MS = 5000

/** Exponential backoff: 500ms, 1s, 2s, 4s, then 5s forever. */
export function reconnectDelayMs (attempt) {
  return Math.min(RECONNECT_MAX_MS, RECONNECT_BASE_MS * 2 ** Math.max(0, attempt))
}

/** Show or hide a small fixed banner (plain DOM, independent of Phaser scenes). */
export function setDisconnectedBanner (visible, text = 'Disconnected — reconnecting…') {
  if (typeof document === 'undefined') {
    return
  }
  let el = document.getElementById('greyvar-disconnected')
  if (!visible) {
    el?.remove()
    return
  }
  if (el == null) {
    el = document.createElement('div')
    el.id = 'greyvar-disconnected'
    el.setAttribute('role', 'status')
    Object.assign(el.style, {
      position: 'fixed',
      top: '12px',
      left: '50%',
      transform: 'translateX(-50%)',
      padding: '6px 14px',
      background: 'rgba(20, 18, 24, 0.85)',
      color: '#f3d36b',
      font: '14px monospace',
      borderRadius: '6px',
      zIndex: 1000,
      pointerEvents: 'none'
    })
    document.body.appendChild(el)
  }
  el.textContent = text
}
