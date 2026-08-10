import type { VolunteerForm } from './volunteer-store.svelte';

const fieldCountWithoutDeparture = 6;
const maximumFieldCount = 9;

export function parseSheetVolunteer(raw: string): Omit<VolunteerForm, 'avatar_url'> {
	const fields = sheetFields(raw);
	if (fields.length < fieldCountWithoutDeparture || fields.length > maximumFieldCount) {
		throw new Error('Dữ liệu cần gồm từ 6 đến 9 ô theo đúng thứ tự');
	}

	const [
		fullName,
		dharmaName,
		birthDate,
		cultivationPlace,
		phone,
		arrivalDate,
		departureDate = '',
		department = '',
		notes = ''
	] = fields;
	if (!fullName) throw new Error('Họ tên không được để trống');
	if (Array.from(department).length > 60) {
		throw new Error('Phân ban không được vượt quá 60 ký tự');
	}

	return {
		full_name: fullName,
		dharma_name: dharmaName,
		birth_date: birthDate,
		cultivation_place: cultivationPlace,
		phone,
		department,
		notes,
		arrival_date: parseDate(arrivalDate, 'Ngày đến'),
		departure_date: departureDate ? parseDate(departureDate, 'Ngày ra về') : ''
	};
}

function sheetFields(raw: string): string[] {
	const normalized = raw.replace(/\r\n?/g, '\n');
	const separator = normalized.includes('\t') ? /\t/ : /\n/;
	const source = normalized.includes('\t')
		? (normalized.split('\n').find((row) => row.trim() !== '') ?? '')
		: normalized;
	const fields = source.split(separator).map((value) => value.trim());
	while (fields.length > 0 && fields[0] === '') fields.shift();
	while (fields.length > fieldCountWithoutDeparture && fields.at(-1) === '') fields.pop();
	return fields;
}

function parseDate(value: string, label: string): string {
	const trimmed = value.trim();
	if (/^\d{4}-\d{2}-\d{2}$/.test(trimmed) && validISODate(trimmed)) return trimmed;

	const match = /^(\d{1,2})[/.\-](\d{1,2})[/.\-](\d{4})$/.exec(trimmed);
	if (!match) throw new Error(`${label} không đúng định dạng ngày`);
	const iso = `${match[3]}-${match[2].padStart(2, '0')}-${match[1].padStart(2, '0')}`;
	if (!validISODate(iso)) throw new Error(`${label} không hợp lệ`);
	return iso;
}

function validISODate(value: string): boolean {
	const [year, month, day] = value.split('-').map(Number);
	const date = new Date(Date.UTC(year, month - 1, day));
	return (
		date.getUTCFullYear() === year && date.getUTCMonth() === month - 1 && date.getUTCDate() === day
	);
}
