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
		arrival_date: parseSheetDate(arrivalDate, 'Ngày đến'),
		departure_date: departureDate ? parseSheetDate(departureDate, 'Ngày ra về') : ''
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

export function parseSheetDate(value: string, label = 'Ngày'): string {
	const input = value
		.trim()
		.replace(/\u00a0/g, ' ')
		.replace(/\s+/g, ' ');
	if (!input) throw new Error(`${label} không được để trống`);

	const isoDateTime = /^(\d{4})-(\d{1,2})-(\d{1,2})(?:[T\s].*)?$/.exec(input);
	if (isoDateTime) return checkedDate(isoDateTime[1], isoDateTime[2], isoDateTime[3], label);

	const compact = /^((?:19|20)\d{2})(\d{2})(\d{2})$/.exec(input);
	if (compact) return checkedDate(compact[1], compact[2], compact[3], label);
	const compactDMY = /^(\d{2})(\d{2})(\d{4})$/.exec(input);
	if (compactDMY) return checkedDate(compactDMY[3], compactDMY[2], compactDMY[1], label);

	const numeric =
		/^(\d{1,4})[/\.\-](\d{1,2})[/\.\-](\d{2}|\d{4})(?:\s+\d{1,2}:\d{2}(?::\d{2})?(?:\s*[ap]m)?)?$/i.exec(
			input
		);
	if (numeric) {
		if (numeric[1].length === 4) {
			return checkedDate(numeric[1], numeric[2], numeric[3], label);
		}
		const first = Number(numeric[1]);
		const second = Number(numeric[2]);
		const year = fourDigitYear(numeric[3]);
		// DMY is the default. Switch to US MDY only when DMY is impossible.
		return second > 12 && first <= 12
			? checkedDate(year, numeric[1], numeric[2], label)
			: checkedDate(year, numeric[2], numeric[1], label);
	}

	const normalizedText = normalizeDateText(input);
	const dayFirstText =
		/^(\d{1,2})\s+(?:(?:thang|thg)\s+)?([a-z]+|\d{1,2})\s+(?:nam\s+)?(\d{2}|\d{4})$/.exec(
			normalizedText
		);
	if (dayFirstText) {
		const month = monthNumber(dayFirstText[2]);
		if (month) return checkedDate(fourDigitYear(dayFirstText[3]), month, dayFirstText[1], label);
	}
	const monthFirstText = /^([a-z]+)\s+(\d{1,2})\s+(\d{2}|\d{4})$/.exec(normalizedText);
	if (monthFirstText) {
		const month = monthNumber(monthFirstText[1]);
		if (month)
			return checkedDate(fourDigitYear(monthFirstText[3]), month, monthFirstText[2], label);
	}
	const vietnameseMonthFirst = /^(?:thang|thg)\s+(\d{1,2})\s+(\d{1,2})\s+(\d{2}|\d{4})$/.exec(
		normalizedText
	);
	if (vietnameseMonthFirst) {
		return checkedDate(
			fourDigitYear(vietnameseMonthFirst[3]),
			vietnameseMonthFirst[1],
			vietnameseMonthFirst[2],
			label
		);
	}

	const serial = /^(\d{5})(?:\.\d+)?$/.exec(input);
	if (serial) {
		const days = Number(serial[1]);
		if (days >= 20_000 && days <= 80_000) {
			const date = new Date(Date.UTC(1899, 11, 30) + days * 86_400_000);
			return formatISO(date.getUTCFullYear(), date.getUTCMonth() + 1, date.getUTCDate());
		}
	}

	throw new Error(`${label} không đúng định dạng ngày`);
}

function checkedDate(
	yearValue: string | number,
	monthValue: string | number,
	dayValue: string | number,
	label: string
): string {
	const year = Number(yearValue);
	const month = Number(monthValue);
	const day = Number(dayValue);
	const iso = formatISO(year, month, day);
	if (!validISODate(iso)) throw new Error(`${label} không hợp lệ`);
	return iso;
}

function formatISO(year: number, month: number, day: number): string {
	return `${String(year).padStart(4, '0')}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
}

function fourDigitYear(value: string): string {
	if (value.length === 4) return value;
	const year = Number(value);
	return String(year >= 80 ? 1900 + year : 2000 + year);
}

function normalizeDateText(value: string): string {
	return value
		.normalize('NFD')
		.replace(/[\u0300-\u036f]/g, '')
		.toLowerCase()
		.replace(/[,/\.\-]/g, ' ')
		.replace(/\s+/g, ' ')
		.trim();
}

function monthNumber(value: string): number {
	if (/^\d{1,2}$/.test(value)) return Number(value);
	return (
		{
			jan: 1,
			january: 1,
			feb: 2,
			february: 2,
			mar: 3,
			march: 3,
			apr: 4,
			april: 4,
			may: 5,
			jun: 6,
			june: 6,
			jul: 7,
			july: 7,
			aug: 8,
			august: 8,
			sep: 9,
			sept: 9,
			september: 9,
			oct: 10,
			october: 10,
			nov: 11,
			november: 11,
			dec: 12,
			december: 12
		}[value] ?? 0
	);
}

function validISODate(value: string): boolean {
	const [year, month, day] = value.split('-').map(Number);
	const date = new Date(Date.UTC(year, month - 1, day));
	return (
		date.getUTCFullYear() === year && date.getUTCMonth() === month - 1 && date.getUTCDate() === day
	);
}
