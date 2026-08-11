const locks = new Set<symbol>();
let originalOverflow = '';

export function setScrollLock(id: symbol, enabled: boolean) {
	if (typeof document === 'undefined') return;
	if (enabled) {
		if (locks.has(id)) return;
		if (locks.size === 0) {
			originalOverflow = document.documentElement.style.overflow;
			document.documentElement.style.overflow = 'hidden';
		}
		locks.add(id);
		return;
	}
	locks.delete(id);
	if (locks.size === 0) document.documentElement.style.overflow = originalOverflow;
}
