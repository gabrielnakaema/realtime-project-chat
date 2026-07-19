const MIN_CHAR_CODE = 0x21; // '!'
const MAX_CHAR_CODE = 0x7e; // '~'

export function generateKeyBetween(a: string, b: string): string {
  if (a !== '' && b !== '' && a >= b) {
    throw new Error(`invalid bounds: "${a}" >= "${b}"`);
  }
  if (a === '' && b === '') return midChar(MIN_CHAR_CODE, MAX_CHAR_CODE);
  if (a === '') return before(b);
  if (b === '') return after(a);
  return between(a, b);
}

function between(a: string, b: string): string {
  const n = Math.min(a.length, b.length);
  for (let i = 0; i < n; i++) {
    const ca = a.charCodeAt(i);
    const cb = b.charCodeAt(i);
    if (ca !== cb) {
      if (cb - ca > 1) {
        return a.slice(0, i) + midChar(ca, cb);
      }
      return a.slice(0, i + 1) + after(a.slice(i + 1));
    }
  }
  return a + before(b.slice(a.length));
}

function before(b: string): string {
  for (let i = 0; i < b.length; i++) {
    const c = b.charCodeAt(i);
    if (c > MIN_CHAR_CODE) {
      const mid = midCode(MIN_CHAR_CODE, c);
      if (mid > MIN_CHAR_CODE) {
        return b.slice(0, i) + String.fromCharCode(mid);
      }
      return b.slice(0, i) + String.fromCharCode(MIN_CHAR_CODE) + midChar(MIN_CHAR_CODE, MAX_CHAR_CODE);
    }
  }
  return b + midChar(MIN_CHAR_CODE, MAX_CHAR_CODE);
}

function after(a: string): string {
  for (let i = a.length - 1; i >= 0; i--) {
    const c = a.charCodeAt(i);
    if (c < MAX_CHAR_CODE) {
      return a.slice(0, i) + midChar(c, MAX_CHAR_CODE);
    }
  }
  return a + midChar(MIN_CHAR_CODE, MAX_CHAR_CODE);
}

function midCode(lo: number, hi: number): number {
  return lo + Math.floor((hi - lo) / 2);
}

function midChar(lo: number, hi: number): string {
  return String.fromCharCode(midCode(lo, hi));
}
