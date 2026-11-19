function itemTitle (itemId) {
  const def = window.gameState?.itemDefinitions?.get(itemId)
  return def?.title ?? itemId
}

function recipeLabel (recipe) {
  const inputs = (recipe.inputs ?? [])
    .map((input) => `${itemTitle(input.item)} x${input.count ?? 1}`)
    .join(' + ')
  const outputs = (recipe.outputs ?? [])
    .map((output) => `${itemTitle(output.item)} x${output.count ?? 1}`)
    .join(', ')
  const title = recipe.title ?? recipe.id
  return `${title}: ${inputs} → ${outputs}`
}

function canCraftRecipe (recipe) {
  const inventory = window.gameState?.inventory ?? []
  const counts = new Map(inventory.map((slot) => [slot.definition, slot.count ?? 0]))

  for (const input of recipe.inputs ?? []) {
    const need = input.count ?? 1
    if ((counts.get(input.item) ?? 0) < need) {
      return false
    }
  }

  return true
}

export function buildInventoryMenuItems () {
  const items = window.gameState?.inventory ?? []
  if (items.length === 0) {
    return [{
      id: 'empty',
      label: '(empty)',
      enabled: false
    }]
  }

  // Choosing an item holds it (drawn in the player's hand); choosing the held
  // item puts it away. The server checks the item is really in the inventory.
  const held = localHeldItem()
  return items.map((slot, index) => {
    const isHeld = slot.definition === held
    return {
      id: `inv-${slot.definition}-${index}`,
      label: `${itemTitle(slot.definition)} x${slot.count ?? 1}${isHeld ? '  (held)' : ''}`,
      enabled: true,
      action: () => {
        window.serverConnection?.sendAppearanceRequest({ heldItem: isHeld ? '-' : slot.definition })
        window.menuSystem?.close()
      }
    }
  })
}

/** Item the local player is holding, from its sprite's current appearance. */
export function localHeldItem () {
  const gs = window.gameState
  const ent = gs?.gridScene?.entities?.[gs?.localPlayerEntityId]
  return ent?.img?.appearance?.heldItem || ent?.appearance?.heldItem || ''
}

export function buildCraftingMenuItems () {
  const recipes = window.gameState?.recipeDefinitions ?? []
  if (recipes.length === 0) {
    return [{
      id: 'no-recipes',
      label: '(no recipes)',
      enabled: false
    }]
  }

  return recipes.map((recipe) => ({
    id: `craft-${recipe.id}`,
    label: recipeLabel(recipe),
    enabled: canCraftRecipe(recipe),
    action: () => {
      window.serverConnection?.sendCraftRequest(recipe.id)
      window.menuSystem?.close()
    }
  }))
}

function menuBlocked () {
  return window.gameConsole?.isOpen?.()
}

export function openInventoryMenu () {
  if (menuBlocked()) {
    return
  }

  const ms = window.menuSystem
  if (ms == null) {
    return
  }

  ms.registerMenu({
    id: 'inventory',
    title: 'Inventory',
    subtitle: 'Collected items',
    items: buildInventoryMenuItems()
  })
  ms.open('inventory')
}

export function openCraftingMenu () {
  if (menuBlocked()) {
    return
  }

  const ms = window.menuSystem
  if (ms == null) {
    return
  }

  ms.registerMenu({
    id: 'crafting',
    title: 'Crafting',
    subtitle: 'Combine materials',
    items: buildCraftingMenuItems()
  })
  ms.open('crafting')
}

export function registerInventoryMenus (menuSystem) {
  menuSystem.registerMenu({
    id: 'inventory',
    title: 'Inventory',
    subtitle: 'Collected items',
    items: []
  })

  menuSystem.registerMenu({
    id: 'crafting',
    title: 'Crafting',
    subtitle: 'Combine materials',
    items: []
  })

  return {
    openInventory: () => openInventoryMenu(),
    openCrafting: () => openCraftingMenu()
  }
}
