export function toDateTimeLocal(value: string): string {
	const date = new Date(value);
	const offset = date.getTimezoneOffset() * 60_000;
	return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}

export function fromDateTimeLocal(value: string): string {
	return new Date(value).toISOString();
}

export function formatEventDate(value: string): string {
	return new Intl.DateTimeFormat(undefined, {
		dateStyle: 'medium',
		timeStyle: 'short'
	}).format(new Date(value));
}

export function formatReminderOffset(minutes: number): string {
	if (minutes === 0) return 'Đúng thời điểm sự kiện';
	if (minutes < 60) return `Trước ${minutes} phút`;

	const hours = minutes / 60;
	if (Number.isInteger(hours) && hours < 24) {
		return `Trước ${hours} giờ`;
	}

	const days = minutes / 1440;
	if (Number.isInteger(days)) {
		return `Trước ${days} ngày`;
	}

	return `Trước ${minutes} phút`;
}
