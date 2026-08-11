import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { parseSheetDate, parseSheetVolunteer } from './sheet-parser.ts';

describe('parseSheetDate', () => {
	const cases = [
		['14/08/2026', '2026-08-14'],
		['14-8-2026', '2026-08-14'],
		['14.08.26', '2026-08-14'],
		['14/08/2026 08:30:00', '2026-08-14'],
		['8/14/2026 8:30 PM', '2026-08-14'],
		['2026-8-14', '2026-08-14'],
		['2026/08/14', '2026-08-14'],
		['2026-08-14 00:00:00', '2026-08-14'],
		['2026-08-14T00:00:00.000Z', '2026-08-14'],
		['20260814', '2026-08-14'],
		['14082026', '2026-08-14'],
		['8/14/2026', '2026-08-14'],
		['14 Aug 2026', '2026-08-14'],
		['August 14, 2026', '2026-08-14'],
		['14 thg 8, 2026', '2026-08-14'],
		['14 tháng 8 năm 2026', '2026-08-14'],
		['Thg 8 14, 2026', '2026-08-14'],
		['46248', '2026-08-14']
	] as const;

	for (const [input, expected] of cases) {
		it(`parses ${input}`, () => assert.equal(parseSheetDate(input), expected));
	}

	it('uses Vietnamese DMY for ambiguous numeric dates', () => {
		assert.equal(parseSheetDate('8/9/2026'), '2026-09-08');
	});

	it('rejects invalid calendar dates', () => {
		assert.throws(() => parseSheetDate('31/02/2026'), /không hợp lệ/);
	});
});

describe('parseSheetVolunteer', () => {
	it('normalizes operational dates and preserves free-form birth date', () => {
		const parsed = parseSheetVolunteer(
			[
				'NGUYỄN VĂN A',
				'TRUNG TIẾN DUY',
				'Khoảng năm 2002',
				'CNT HÀ NỘI',
				'0388888666',
				'Aug 14, 2026',
				'14.12.2026'
			].join('\n')
		);

		assert.equal(parsed.birth_date, 'Khoảng năm 2002');
		assert.equal(parsed.arrival_date, '2026-08-14');
		assert.equal(parsed.departure_date, '2026-12-14');
	});
});
