<script module lang="ts">
	function formatDate(value: string) {
		return new Intl.DateTimeFormat('vi-VN').format(new Date(value));
	}
</script>

<script lang="ts">
	import { onMount } from 'svelte';
	import { router } from '$lib/navigation/router.svelte';
	import { volunteerStore } from '$lib/volunteers/volunteer-store.svelte';
	import { volunteerStatus } from '$lib/volunteers/status';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';
	import { popupStore } from '$lib/ui/popup-store.svelte';
	import AvatarUploader from '$lib/ui/AvatarUploader.svelte';
	import { uploadAvatar } from '$lib/uploads/api';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';

	let { volunteerId }: { volunteerId: string } = $props();
	let item = $derived(volunteerStore.selected?.id === volunteerId ? volunteerStore.selected : null);
	let loading = $state(true);
	let clock = $state(Date.now());
	let status = $derived(volunteerStatus(item?.departure_date, new Date(clock)));
	let avatarUploading = $state(false);

	onMount(() => {
		void volunteerStore.fetch(volunteerId).finally(() => (loading = false));
		const timer = window.setInterval(() => (clock = Date.now()), 60_000);
		return () => window.clearInterval(timer);
	});

	async function remove() {
		const accepted = await popupStore.confirm({
			title: 'Xoá Huynh đệ',
			message: 'Huynh đệ này sẽ bị xoá vĩnh viễn.',
			confirmLabel: 'Xoá',
			cancelLabel: 'Huỷ',
			tone: 'danger'
		});
		if (accepted && (await volunteerStore.remove(volunteerId))) router.replace('/volunteers');
	}

	async function selectAvatar(file: File) {
		avatarUploading = true;
		try {
			const avatarURL = await uploadAvatar(file);
			await volunteerStore.updateAvatar(volunteerId, avatarURL);
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể tải ảnh lên');
		} finally {
			avatarUploading = false;
		}
	}
</script>

<section class="h-full overflow-y-auto px-4 py-4 md:px-6 md:py-6 lg:px-8">
	<div class="mx-auto max-w-5xl">
		{#if loading}
			<div class="py-16"><LoadingIndicator label="Đang tải Huynh đệ..." /></div>
		{:else if item}
			<div class="border-b border-[var(--color-border)] pb-5">
				<AvatarUploader
					avatarUrl={item.avatar_url}
					displayName={item.full_name}
					editable={authStore.can('volunteer.update')}
					uploading={avatarUploading || volunteerStore.isAvatarSaving}
					onselect={selectAvatar}
				/>
				<div class="mt-5 min-w-0">
					<h2 class="truncate text-xl font-semibold">{item.full_name}</h2>
					<p class="mt-1 text-sm text-[var(--color-text-secondary)]">
						{item.dharma_name || 'Chưa có pháp danh'}
					</p>
					<span
						class={[
							'mt-2 inline-flex rounded-full px-2.5 py-1 text-xs font-medium',
							status === 'departed'
								? 'bg-[var(--color-surface-muted)] text-[var(--color-text-secondary)]'
								: 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
						]}
					>
						{status === 'departed' ? 'Đã ra về' : 'Đang công quả'}
					</span>
				</div>
			</div>

			<dl class="grid border-b border-[var(--color-border)] md:grid-cols-2 md:gap-x-8">
				{@render row('Ngày sinh', item.birth_date || 'Chưa cập nhật')}
				{@render row('Nơi sinh hoạt', item.cultivation_place || 'Chưa cập nhật')}
				{@render row('Số điện thoại', item.phone || 'Chưa cập nhật')}
				{@render row('Phân ban', item.department || 'Chưa cập nhật')}
				{@render row('Ngày đến', formatDate(item.arrival_date))}
				{@render row(
					'Ngày ra về',
					item.departure_date ? formatDate(item.departure_date) : 'Chưa xác định'
				)}
			</dl>
			{#if item.notes}
				<div class="border-t border-[var(--color-border)] py-4">
					<p class="text-sm text-[var(--color-text-secondary)]">Ghi chú</p>
					<p class="mt-2 text-sm leading-6 whitespace-pre-wrap">{item.notes}</p>
				</div>
			{/if}

			{#if authStore.can('volunteer.update') || authStore.can('volunteer.delete')}
				<div class="mt-6 flex justify-end gap-3">
					{#if authStore.can('volunteer.update')}<button
							type="button"
							class="flex h-11 min-w-36 items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] px-4 text-sm font-semibold text-white"
							onclick={() => router.push(`/volunteers/${item.id}/edit`)}
						>
							<span class="icon-[lucide--pencil] h-4 w-4" aria-hidden="true"></span>Sửa
						</button>{/if}
					{#if authStore.can('volunteer.delete')}<button
							type="button"
							class="flex h-11 min-w-36 items-center justify-center gap-2 rounded-md border border-[var(--color-danger)] px-4 text-sm font-semibold text-[var(--color-danger)]"
							onclick={remove}
						>
							<span class="icon-[lucide--trash-2] h-4 w-4" aria-hidden="true"></span>Xoá
						</button>{/if}
				</div>
			{/if}
		{/if}
	</div>
</section>

{#snippet row(label: string, value: string)}
	<div class="grid grid-cols-[8rem_1fr] gap-3 border-t border-[var(--color-border)] py-3.5 text-sm">
		<dt class="text-[var(--color-text-secondary)]">{label}</dt>
		<dd class="text-right font-medium break-words">{value}</dd>
	</div>
{/snippet}
