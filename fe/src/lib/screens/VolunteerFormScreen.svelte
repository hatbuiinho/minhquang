<script lang="ts">
	import { tick } from 'svelte';
	import { router } from '$lib/navigation/router.svelte';
	import { volunteerStore } from '$lib/volunteers/volunteer-store.svelte';
	import { parseSheetVolunteer } from '$lib/volunteers/sheet-parser';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import DepartmentCombobox from '$lib/volunteers/DepartmentCombobox.svelte';

	let { volunteerId }: { volunteerId?: string } = $props();
	let sheetData = $state('');
	let sheetOpen = $state(false);
	let sheetTextarea = $state<HTMLTextAreaElement>();

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		const item = await volunteerStore.save(volunteerId);
		if (!item) return;
		router.replace(volunteerId ? `/volunteers/${item.id}` : '/volunteers');
	}

	function applySheetData() {
		try {
			const parsed = parseSheetVolunteer(sheetData);
			volunteerStore.form = { ...volunteerStore.form, ...parsed };
			sheetData = '';
			sheetOpen = false;
			toastStore.success('Đã điền dữ liệu từ Google Sheets');
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể đọc dữ liệu');
		}
	}

	async function toggleSheetInput() {
		sheetOpen = !sheetOpen;
		if (!sheetOpen) return;

		await tick();
		sheetTextarea?.focus();
	}
</script>

<form
	id="volunteer-form"
	class="h-full overflow-y-auto px-4 py-4 md:px-6 md:py-6 lg:px-8"
	onsubmit={submit}
>
	<div class="mx-auto max-w-5xl pb-8">
		<div class="border-b border-[var(--color-border)] pb-4">
			<button
				type="button"
				class="flex h-11 w-full items-center justify-center gap-2 rounded-md border border-[var(--color-primary)] bg-[var(--color-primary-soft)] text-sm font-semibold text-[var(--color-primary-dark)]"
				onclick={toggleSheetInput}
			>
				<span class="icon-[lucide--clipboard-paste] h-5 w-5" aria-hidden="true"></span>
				Dán dữ liệu Google Sheets
			</button>
			{#if sheetOpen}
				<div class="mt-3 space-y-2">
					<label class="block">
						<span class="mb-1.5 block text-sm font-medium">Dữ liệu từ Google Sheets</span>
						<textarea
							bind:this={sheetTextarea}
							bind:value={sheetData}
							rows="7"
							class="w-full resize-y rounded-md border-[var(--color-border-strong)] text-sm"
						></textarea>
					</label>
					<button
						type="button"
						disabled={!sheetData.trim()}
						class="h-10 w-full rounded-md bg-[var(--color-primary)] text-sm font-semibold text-white disabled:opacity-50"
						onclick={applySheetData}
					>
						Điền dữ liệu
					</button>
				</div>
			{/if}
		</div>
		<div class="mt-5 grid gap-4 md:grid-cols-2 md:gap-x-6 md:gap-y-5">
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium"
					>Họ tên <span class="text-[var(--color-danger)]">*</span></span
				>
				<input
					bind:value={volunteerStore.form.full_name}
					required
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				/>
			</label>
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium">Pháp danh</span>
				<input
					bind:value={volunteerStore.form.dharma_name}
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				/>
			</label>
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium">Ngày sinh</span>
				<input
					bind:value={volunteerStore.form.birth_date}
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				/>
			</label>
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium">Số điện thoại</span>
				<input
					bind:value={volunteerStore.form.phone}
					type="tel"
					inputmode="tel"
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				/>
			</label>
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium">Nơi sinh hoạt</span>
				<input
					bind:value={volunteerStore.form.cultivation_place}
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				/>
			</label>
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium">Phân ban</span>
				<DepartmentCombobox bind:value={volunteerStore.form.department} />
			</label>
			<label class="block md:col-span-2">
				<span class="mb-1.5 block text-sm font-medium">Ảnh đại diện (URL)</span>
				<input
					bind:value={volunteerStore.form.avatar_url}
					type="url"
					placeholder="https://..."
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				/>
			</label>
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium"
					>Ngày đến <span class="text-[var(--color-danger)]">*</span></span
				>
				<input
					bind:value={volunteerStore.form.arrival_date}
					type="date"
					required
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				/>
			</label>
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium">Ngày ra về</span>
				<input
					bind:value={volunteerStore.form.departure_date}
					type="date"
					min={volunteerStore.form.arrival_date}
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				/>
			</label>
			<label class="block md:col-span-2">
				<span class="mb-1.5 block text-sm font-medium">Ghi chú</span>
				<textarea
					bind:value={volunteerStore.form.notes}
					rows="4"
					class="w-full resize-y rounded-md border-[var(--color-border-strong)]"></textarea>
			</label>
		</div>
	</div>
</form>
