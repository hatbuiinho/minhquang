<script lang="ts">
	import { eventStore } from '$lib/events/event-store.svelte';
	import { formatEventDate, formatReminderOffset } from '$lib/events/time';
	import { router } from '$lib/navigation/router.svelte';
	import { popupStore } from '$lib/ui/popup-store.svelte';

	let { eventId }: { eventId?: string } = $props();
	let event = $derived(eventStore.get(eventId));

	$effect(() => {
		if (!event && eventId) {
			void eventStore.loadById(eventId);
		}
	});

	async function remove() {
		if (!event) return;

		const confirmed = await popupStore.confirm({
			title: 'Xoá sự kiện',
			message: 'Bạn có chắc muốn xoá sự kiện này không? Hành động này không thể hoàn tác.',
			confirmLabel: 'Xoá',
			cancelLabel: 'Huỷ',
			tone: 'danger'
		});
		if (!confirmed) return;

		const removed = await eventStore.remove(event.id);
		if (removed) {
			router.replace('/events');
		}
	}
</script>

<section class="h-full overflow-y-auto px-4 py-4">
	{#if !event && (eventStore.isEventLoading(eventId) || eventStore.isLoading)}
		<div
			class="rounded-[14px] border border-[var(--color-border)] bg-[var(--color-surface)] p-5 text-center text-sm text-[var(--color-text-secondary)]"
		>
			Đang tải sự kiện...
		</div>
	{:else if event}
		<div
			class="rounded-[16px] border border-[var(--color-border)] bg-[var(--color-surface)] p-4 shadow-[var(--shadow-soft)]"
		>
			<p class="text-xs font-semibold text-[var(--color-primary)] uppercase">{event.status}</p>
			<h2 class="mt-2 text-2xl font-semibold text-[var(--color-text)]">{event.title}</h2>

			<dl class="mt-5 space-y-4">
				<div>
					<dt class="text-xs font-semibold text-[var(--color-text-muted)] uppercase">Thời gian</dt>
					<dd class="mt-1 text-sm text-[var(--color-text)]">{formatEventDate(event.starts_at)}</dd>
				</div>
				<div>
					<dt class="text-xs font-semibold text-[var(--color-text-muted)] uppercase">Múi giờ</dt>
					<dd class="mt-1 text-sm text-[var(--color-text)]">{event.timezone}</dd>
				</div>
				<div>
					<dt class="text-xs font-semibold text-[var(--color-text-muted)] uppercase">Mô tả</dt>
					<dd class="mt-1 text-sm leading-6 whitespace-pre-wrap text-[var(--color-text-secondary)]">
						{event.description || 'Chưa có mô tả.'}
					</dd>
				</div>
				<div>
					<dt class="text-xs font-semibold text-[var(--color-text-muted)] uppercase">Nhắc hẹn</dt>
					<dd class="mt-2">
						{#if event.reminders.length === 0}
							<span class="text-sm text-[var(--color-text-secondary)]">Chưa có nhắc hẹn.</span>
						{:else}
							<ul class="space-y-2">
								{#each event.reminders as reminder (reminder.id)}
									<li
										class="rounded-[12px] bg-[var(--color-primary-soft)] px-3 py-2 text-sm text-[var(--color-primary-dark)]"
									>
										{formatReminderOffset(reminder.offset_minutes)}
										{#if !reminder.enabled}
											<span class="text-[var(--color-text-secondary)]"> · đã tắt</span>
										{/if}
									</li>
								{/each}
							</ul>
						{/if}
					</dd>
				</div>
			</dl>
		</div>

		<div class="mt-4 grid grid-cols-2 gap-3">
			<button
				type="button"
				class="flex h-11 items-center justify-center gap-2 rounded-[12px] border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-sm font-semibold text-[var(--color-text)]"
				onclick={() => router.push(`/events/${encodeURIComponent(event.id)}/edit`)}
			>
				<span class="icon-[lucide--pencil] h-4 w-4" aria-hidden="true"></span>
				Sửa
			</button>
			<button
				type="button"
				class="flex h-11 items-center justify-center gap-2 rounded-[12px] border border-[var(--color-danger)] bg-[var(--color-danger-soft)] text-sm font-semibold text-[var(--color-danger)]"
				onclick={remove}
			>
				<span class="icon-[lucide--trash-2] h-4 w-4" aria-hidden="true"></span>
				Xoá
			</button>
		</div>
	{:else}
		<div
			class="rounded-[14px] border border-[var(--color-border)] bg-[var(--color-surface)] p-5 text-center text-sm text-[var(--color-text-secondary)]"
		>
			Không tìm thấy sự kiện.
		</div>
	{/if}
</section>
