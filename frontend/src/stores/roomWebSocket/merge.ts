export function mergeById<T, K>(existing: T[], incoming: T[], getId: (item: T) => K): T[] {
  return [...new Map([...existing, ...incoming].map((item) => [getId(item), item])).values()]
}
