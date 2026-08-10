<script lang="ts">
	import { onMount } from 'svelte';

	import {
		dismissReminderJob,
		listUpcomingReminders,
		listReminderInbox,
		markReminderJobRead,
		snoozeReminderJob,
		type ReminderJob
	} from '$lib/events/api';
	import { formatEventDate, formatReminderOffset } from '$lib/events/time';
	import { router } from '$lib/navigation/router.svelte';
	import BottomSheet from '$lib/ui/BottomSheet.svelte';
	import { popupStore } from '$lib/ui/popup-store.svelte';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';

	type ReminderView = 'inbox' | 'upcoming';
	type SnoozeOption = {
		label: string;
		delayMinutes: number;
	};

	const snoozeOptions: SnoozeOption[] = [
		{ label: '15 phút', delayMinutes: 15 },
		{ label: '1 giờ', delayMinutes: 60 },
		{ label: 'Ngày mai', delayMinutes: 24 * 60 }
	];
	const remindersRefreshTTL = 20_000;

	let { active = false, refreshKey = 0 }: { active?: boolean; refreshKey?: number } = $props();
	let view = $state<ReminderView>('upcoming');
	let inboxReminders = $state<ReminderJob[]>([]);
	let upcomingReminders = $state<ReminderJob[]>([]);
	let isLoading = $state(false);
	let activeReminderId = $state('');
	let snoozeTarget = $state<ReminderJob | undefined>();
	let error = $state('');
	let lastLoadedAt = $state(0);
	let reminders = $derived(view === 'inbox' ? inboxReminders : upcomingReminders);
	let pastCount = $derived(inboxReminders.length);

	onMount(() => {
		void load();
	});

	$effect(() => {
		if (!active) return;

		Number(refreshKey);
		void refreshIfStale();
	});

	async function load() {
		if (isLoading) return;

		isLoading = true;
		error = '';
		try {
			const [inbox, upcoming] = await Promise.all([listReminderInbox(), listUpcomingReminders()]);
			inboxReminders = inbox;
			upcomingReminders = upcoming;
			lastLoadedAt = Date.now();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Không tải được nhắc hẹn';
		} finally {
			isLoading = false;
		}
	}

	async function refreshIfStale() {
		if (Date.now() - lastLoadedAt >= remindersRefreshTTL) {
			await load();
		}
	}

	async function openReminder(reminder: ReminderJob) {
		if (!reminder.read_at) {
			activeReminderId = reminder.id;
			try {
				const updated = await markReminderJobRead(reminder.id);
				replaceReminder(updated);
			} catch (err) {
				toastStore.error(err instanceof Error ? err.message : 'Không thể đánh dấu đã đọc');
			} finally {
				activeReminderId = '';
			}
		}

		router.push(`/events/${reminder.event_id}`);
	}

	async function dismiss(reminder: ReminderJob) {
		const isUpcoming = view === 'upcoming';
		const confirmed = await popupStore.confirm({
			title: isUpcoming ? 'Bỏ qua nhắc hẹn' : 'Bỏ qua nhắc hẹn đã qua',
			message: isUpcoming
				? 'Bạn có chắc muốn bỏ qua nhắc hẹn này không? App sẽ không gửi thông báo khi đến hạn.'
				: 'Bạn có chắc muốn bỏ qua nhắc hẹn đã qua này không?',
			confirmLabel: 'Bỏ qua',
			cancelLabel: 'Huỷ',
			tone: 'danger'
		});
		if (!confirmed) return;

		activeReminderId = reminder.id;
		try {
			await dismissReminderJob(reminder.id);
			inboxReminders = inboxReminders.filter((item) => item.id !== reminder.id);
			upcomingReminders = upcomingReminders.filter((item) => item.id !== reminder.id);
			toastStore.success(isUpcoming ? 'Đã bỏ qua nhắc hẹn' : 'Đã bỏ qua nhắc hẹn đã qua');
		} catch (err) {
			toastStore.error(err instanceof Error ? err.message : 'Không thể bỏ qua nhắc hẹn');
		} finally {
			activeReminderId = '';
		}
	}

	async function snooze(option: SnoozeOption) {
		if (!snoozeTarget) return;

		activeReminderId = snoozeTarget.id;
		try {
			const snoozed = await snoozeReminderJob(snoozeTarget.id, option.delayMinutes);
			inboxReminders = inboxReminders.filter((item) => item.id !== snoozeTarget?.id);
			upcomingReminders = [snoozed, ...upcomingReminders]
				.filter(
					(item, index, items) => items.findIndex((candidate) => candidate.id === item.id) === index
				)
				.sort((left, right) => {
					const bySchedule =
						new Date(left.scheduled_at).getTime() - new Date(right.scheduled_at).getTime();
					if (bySchedule !== 0) return bySchedule;
					return left.id.localeCompare(right.id);
				});
			snoozeTarget = undefined;
			toastStore.success(`Đã nhắc lại sau ${option.label.toLowerCase()}`);
		} catch (err) {
			toastStore.error(err instanceof Error ? err.message : 'Không thể nhắc lại');
		} finally {
			activeReminderId = '';
		}
	}

	function replaceReminder(updated: ReminderJob) {
		inboxReminders = inboxReminders.map((item) => (item.id === updated.id ? updated : item));
	}
</script>

<section class="h-full overflow-y-auto px-4 py-4">
	<div class="mb-4 flex items-center justify-between gap-3">
		<div>
			<h2 class="text-lg font-semibold text-[var(--color-text)]">Nhắc hẹn</h2>
			<p class="text-sm text-[var(--color-text-secondary)]">
				{pastCount} đã qua · {upcomingReminders.length} sắp tới
			</p>
		</div>
		<button
			type="button"
			class="grid h-10 w-10 shrink-0 place-items-center rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] text-[var(--color-text)] disabled:opacity-50"
			disabled={isLoading}
			onclick={load}
			aria-label="Làm mới nhắc hẹn"
		>
			<span class="icon-[lucide--refresh-cw] h-5 w-5" aria-hidden="true"></span>
		</button>
	</div>

	<div
		class="mb-4 grid grid-cols-2 rounded-[14px] border border-[var(--color-border)] bg-[var(--color-surface)] p-1"
	>
		<button
			type="button"
			class={[
				'h-10 rounded-[11px] text-sm font-semibold transition',
				view === 'upcoming'
					? 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
					: 'text-[var(--color-text-secondary)]'
			]}
			onclick={() => (view = 'upcoming')}
		>
			Sắp tới
		</button>
		<button
			type="button"
			class={[
				'h-10 rounded-[11px] text-sm font-semibold transition',
				view === 'inbox'
					? 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
					: 'text-[var(--color-text-secondary)]'
			]}
			onclick={() => (view = 'inbox')}
		>
			Đã qua
		</button>
	</div>

	{#if error}
		<div
			class="rounded-[14px] border border-[var(--color-danger)] bg-[var(--color-danger-soft)] p-4 text-sm text-[var(--color-danger)]"
		>
			{error}
		</div>
	{:else if isLoading && reminders.length === 0}
		<div
			class="rounded-[14px] border border-[var(--color-border)] bg-[var(--color-surface)] p-5 text-sm text-[var(--color-text-secondary)]"
		>
			<LoadingIndicator label="Đang tải nhắc hẹn..." />
		</div>
	{:else if reminders.length === 0}
		<div
			class="rounded-[16px] border border-[var(--color-border)] bg-[var(--color-surface)] p-5 shadow-[var(--shadow-soft)]"
		>
			<h3 class="text-base font-semibold text-[var(--color-text)]">
				{view === 'upcoming' ? 'Chưa có nhắc hẹn sắp tới' : 'Chưa có nhắc hẹn đã qua'}
			</h3>
			<p class="mt-2 text-sm leading-6 text-[var(--color-text-secondary)]">
				{view === 'upcoming'
					? 'Các nhắc hẹn trong tương lai của sự kiện đang hoạt động sẽ xuất hiện tại đây.'
					: 'Các nhắc hẹn đã đến hạn hoặc quá hạn sẽ xuất hiện tại đây.'}
			</p>
		</div>
	{:else}
		<ul class="space-y-3">
			{#each reminders as reminder (reminder.id)}
				<li>
					<div
						class={[
							'w-full rounded-[14px] border bg-[var(--color-surface)] p-4 shadow-[var(--shadow-soft)]',
							view === 'inbox' && !reminder.read_at
								? 'border-[var(--color-primary)]'
								: 'border-[var(--color-border)]'
						]}
					>
						<div class="flex items-start gap-3">
							<span
								class="mt-0.5 icon-[lucide--bell-ring] h-5 w-5 shrink-0 text-[var(--color-primary)]"
								aria-hidden="true"
							></span>
							<div class="min-w-0 flex-1">
								<div class="flex items-start justify-between gap-2">
									<p class="truncate text-base font-semibold text-[var(--color-text)]">
										{reminder.event_title}
									</p>
									{#if view === 'inbox' && !reminder.read_at}
										<span
											class="mt-1 h-2 w-2 shrink-0 rounded-full bg-[var(--color-primary)]"
											aria-label="Chưa đọc"
										></span>
									{/if}
								</div>
								<p class="mt-1 text-sm text-[var(--color-text-secondary)]">
									{formatReminderOffset(reminder.offset_minutes)}
								</p>
								<div class="mt-3 grid gap-1 text-xs text-[var(--color-text-secondary)]">
									<span>Nhắc lúc: {formatEventDate(reminder.scheduled_at)}</span>
									<span>Sự kiện: {formatEventDate(reminder.event_starts_at)}</span>
								</div>
								<div
									class={view === 'inbox'
										? 'mt-4 grid grid-cols-3 gap-2'
										: 'mt-4 grid grid-cols-2 gap-2'}
								>
									<button
										type="button"
										class="flex h-10 items-center justify-center gap-2 rounded-[12px] bg-[var(--color-primary-soft)] px-3 text-sm font-semibold text-[var(--color-primary-dark)] disabled:opacity-60"
										disabled={activeReminderId === reminder.id}
										onclick={() =>
											view === 'inbox'
												? openReminder(reminder)
												: router.push(`/events/${reminder.event_id}`)}
									>
										{#if view === 'inbox' && activeReminderId === reminder.id && !reminder.read_at}
											<LoadingIndicator label="Đang mở..." />
										{:else}
											<span class="icon-[lucide--external-link] h-4 w-4" aria-hidden="true"></span>
											Mở sự kiện
										{/if}
									</button>
									{#if view === 'inbox'}
										<button
											type="button"
											class="flex h-10 items-center justify-center gap-1.5 rounded-[12px] border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-2 text-xs font-semibold text-[var(--color-primary)] disabled:opacity-60"
											disabled={activeReminderId === reminder.id}
											onclick={() => (snoozeTarget = reminder)}
										>
											<span class="icon-[lucide--alarm-clock-plus] h-4 w-4" aria-hidden="true"
											></span>
											Nhắc lại
										</button>
										<button
											type="button"
											class="flex h-10 items-center justify-center gap-1.5 rounded-[12px] border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-2 text-xs font-semibold text-[var(--color-text-muted)] disabled:opacity-60"
											disabled={activeReminderId === reminder.id}
											onclick={() => dismiss(reminder)}
										>
											<span class="icon-[lucide--x] h-4 w-4" aria-hidden="true"></span>
											Bỏ qua
										</button>
									{:else}
										<button
											type="button"
											class="flex h-10 items-center justify-center gap-1.5 rounded-[12px] border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-2 text-xs font-semibold text-[var(--color-text-muted)] disabled:opacity-60"
											disabled={activeReminderId === reminder.id}
											onclick={() => dismiss(reminder)}
										>
											<span class="icon-[lucide--x] h-4 w-4" aria-hidden="true"></span>
											Bỏ qua
										</button>
									{/if}
								</div>
							</div>
						</div>
					</div>
				</li>
			{/each}
		</ul>
	{/if}

	<BottomSheet
		open={Boolean(snoozeTarget)}
		title="Nhắc lại sau"
		onClose={() => (snoozeTarget = undefined)}
	>
		<div class="grid gap-2">
			{#each snoozeOptions as option (option.delayMinutes)}
				<button
					type="button"
					class="flex h-12 items-center justify-between rounded-[12px] bg-[var(--color-bg)] px-3 text-left text-sm font-semibold text-[var(--color-text)] disabled:opacity-60"
					disabled={activeReminderId === snoozeTarget?.id}
					onclick={() => snooze(option)}
				>
					<span>{option.label}</span>
					<span
						class="icon-[lucide--chevron-right] h-4 w-4 text-[var(--color-text-muted)]"
						aria-hidden="true"
					></span>
				</button>
			{/each}
		</div>
	</BottomSheet>
</section>
