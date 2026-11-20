// Parses the console `rift` command's arguments into a RiftRequest:
//   rift                         one level above your best, random layout
//   rift 5                       level 5
//   rift 5 seed=7                level 5, a particular layout
//   rift mechanics=lever,keyChest
//   rift leave                   flee the rift you are in
export const RIFT_MECHANICS = ['lever', 'keyChest', 'pushBlock', 'timedSwitch']

export function parseRiftArgs (args) {
  const request = { level: 0, seed: 0, mechanics: [], leave: false }
  const errors = []
  for (const raw of args) {
    const arg = String(raw).trim()
    if (arg === '') {
      continue
    }
    if (arg === 'leave') {
      request.leave = true
      continue
    }
    const eq = arg.indexOf('=')
    if (eq < 0) {
      if (/^\d+$/.test(arg) && Number(arg) > 0) {
        request.level = Number(arg)
      } else {
        errors.push(`not a level: ${arg}`)
      }
      continue
    }
    const key = arg.slice(0, eq)
    const value = arg.slice(eq + 1)
    if (key === 'seed' || key === 'level') {
      if (!/^\d+$/.test(value) || Number(value) <= 0) {
        errors.push(`${key} must be a positive whole number: ${value}`)
      } else {
        request[key] = Number(value)
      }
    } else if (key === 'mechanics') {
      for (const m of value.split(',').map(s => s.trim()).filter(Boolean)) {
        if (!RIFT_MECHANICS.includes(m)) {
          errors.push(`unknown mechanic ${m} (have ${RIFT_MECHANICS.join(', ')})`)
        } else {
          request.mechanics.push(m)
        }
      }
    } else {
      errors.push(`unknown option ${key} (use seed=, mechanics=)`)
    }
  }
  if (request.leave && (request.level > 0 || request.seed > 0 || request.mechanics.length > 0)) {
    errors.push('rift leave takes no other options')
  }
  return { request, errors }
}
