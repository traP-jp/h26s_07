import type { Line } from '@/api/schema'

function cellCenter(index: number) {
  const column = Math.floor(index / 5)
  const row = index % 5

  return {
    x: ((column + 0.5) / 5) * 100,
    y: ((row + 0.5) / 5) * 100,
  }
}

export function createLineEffectStyle(line: Line): Record<string, string> {
  if (!line || line.length < 2) return {}

  const start = cellCenter(line[0]!)
  const end = cellCenter(line[line.length - 1]!)
  const dx = end.x - start.x
  const dy = end.y - start.y
  const length = Math.hypot(dx, dy)
  const angle = Math.atan2(dy, dx)

  return {
    '--line-x': `${start.x}%`,
    '--line-y': `${start.y}%`,
    '--line-length': `${length}%`,
    '--line-angle': `${angle}rad`,
  }
}

export function lineKey(line: Line) {
  return line.join('-')
}

export function isInLine(lines: Line[] | undefined, cellIndex: number) {
  return lines?.some((line) => line.includes(cellIndex)) ?? false
}
