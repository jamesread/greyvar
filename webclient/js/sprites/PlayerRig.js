// Loads layered rigs built by sprite-generator and recolours their sheets.
import Phaser from 'phaser'
import { recolourMap, recolourPixels, recolouredTextureKey } from './appearance.js'

export const PLAYER_RIG_KEY = 'rig:player'
export const PLAYER_RIG_URL = 'img/textures/entities/player/player.rig.json'

/** Queue the rig manifest and, once it arrives, every layer sheet. */
export function preloadRig (scene, key = PLAYER_RIG_KEY, url = PLAYER_RIG_URL) {
  if (scene.cache.json.exists(key)) {
    queueSheets(scene, scene.cache.json.get(key))
    return
  }
  scene.load.once(`filecomplete-json-${key}`, (_k, _t, rig) => queueSheets(scene, rig))
  scene.load.json(key, url)
}

function queueSheets (scene, rig) {
  for (const layer of rig.layers) {
    for (const path of Object.values(layer.sheets)) {
      if (!scene.textures.exists(path)) {
        scene.load.spritesheet(path, path, { frameWidth: rig.frameSize, frameHeight: rig.frameSize })
      }
    }
  }
}

/** The loaded rig, or null until its manifest and sheets are ready. */
export function getRig (scene, key = PLAYER_RIG_KEY) {
  const rig = scene.cache.json.get(key)
  if (rig == null) {
    return null
  }
  for (const layer of rig.layers) {
    for (const path of Object.values(layer.sheets)) {
      if (!scene.textures.exists(path)) {
        return null
      }
    }
  }
  return rig
}

/**
 * Texture key for a layer variant under a palette choice. Recoloured copies
 * are made once per (sheet, relevant palettes) and reused by every player.
 */
export function layerTexture (scene, rig, layer, variant, palettes) {
  const sheet = layer.sheets[variant] ?? layer.sheets[layer.default]
  const key = recolouredTextureKey(sheet, layer.slots, palettes)
  if (scene.textures.exists(key)) {
    return key
  }

  const map = recolourMap(rig, palettes, layer.slots)
  if (map.size === 0) {
    return sheet
  }

  const src = scene.textures.get(sheet).getSourceImage()
  const canvas = document.createElement('canvas')
  canvas.width = src.width
  canvas.height = src.height
  const ctx = canvas.getContext('2d', { willReadFrequently: true })
  ctx.drawImage(src, 0, 0)
  const img = ctx.getImageData(0, 0, canvas.width, canvas.height)
  recolourPixels(img.data, map)
  ctx.putImageData(img, 0, 0)

  const tex = scene.textures.addCanvas(key, canvas)
  const fs = rig.frameSize
  for (let i = 0; i < rig.frameCount; i++) {
    tex.add(i, 0, (i % rig.columns) * fs, Math.floor(i / rig.columns) * fs, fs, fs)
  }
  tex.setFilter(Phaser.Textures.FilterMode.NEAREST)
  return key
}
