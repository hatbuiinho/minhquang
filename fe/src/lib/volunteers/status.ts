export type VolunteerStatus = 'active' | 'departed';

export function volunteerStatus(
	departureDate: string | undefined,
	now = new Date()
): VolunteerStatus {
	if (!departureDate) return 'active';
	return departureDate.slice(0, 10) < vietnamDateKey(now) ? 'departed' : 'active';
}

export function vietnamDateKey(date = new Date()): string {
	const parts = new Intl.DateTimeFormat('en-US', {
		timeZone: 'Asia/Ho_Chi_Minh',
		year: 'numeric',
		month: '2-digit',
		day: '2-digit'
	}).formatToParts(date);
	const values = Object.fromEntries(parts.map((part) => [part.type, part.value]));
	return `${values.year}-${values.month}-${values.day}`;
}
