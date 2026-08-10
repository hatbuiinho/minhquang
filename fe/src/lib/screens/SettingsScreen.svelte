<script lang="ts">
	import { pushNotificationStore } from '$lib/notifications/push-notification-store.svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';
</script>

<section class="h-full overflow-y-auto px-4 py-4">
	<div
		class="rounded-[16px] border border-[var(--color-border)] bg-[var(--color-surface)] p-5 shadow-[var(--shadow-soft)]"
	>
		<h2 class="text-lg font-semibold text-[var(--color-text)]">Cài đặt</h2>
		<div class="mt-4 space-y-3">
			<div class="rounded-[14px] bg-[var(--color-bg)] p-3">
				<p class="text-sm font-medium text-[var(--color-text)]">Hồ sơ</p>
				<p class="mt-1 text-sm text-[var(--color-text-secondary)]">
					Người dùng local trong môi trường dev
				</p>
			</div>
			<div class="rounded-[14px] bg-[var(--color-bg)] p-3">
				<p class="text-sm font-medium text-[var(--color-text)]">Thông báo</p>
				<p class="mt-1 text-sm text-[var(--color-text-secondary)]">
					Đăng ký thiết bị với Firebase Cloud Messaging để nhận nhắc hẹn.
				</p>
				<button
					type="button"
					class="mt-3 flex h-10 w-full items-center justify-center gap-2 rounded-[12px] bg-[var(--color-primary)] text-sm font-semibold text-white disabled:opacity-60"
					disabled={pushNotificationStore.status === 'requesting'}
					onclick={() => pushNotificationStore.register()}
				>
					{#if pushNotificationStore.status === 'requesting'}
						<LoadingIndicator label="Đang đăng ký..." />
					{:else}
						<span class="icon-[lucide--bell-plus] h-4 w-4" aria-hidden="true"></span>
						Bật thông báo
					{/if}
				</button>

				{#if pushNotificationStore.error}
					<p
						class="mt-3 rounded-[12px] border border-[var(--color-danger)] bg-[var(--color-danger-soft)] px-3 py-2 text-sm text-[var(--color-danger)]"
					>
						{pushNotificationStore.error}
					</p>
				{/if}

				{#if pushNotificationStore.deviceSyncStatus === 'syncing'}
					<div class="mt-3">
						<LoadingIndicator label="Đang lưu thiết bị..." />
					</div>
				{:else if pushNotificationStore.deviceSyncStatus === 'synced'}
					<p
						class="mt-3 rounded-[12px] border border-[var(--color-primary)] bg-[var(--color-primary-soft)] px-3 py-2 text-sm text-[var(--color-primary-dark)]"
					>
						Thiết bị đã được lưu trên backend.
					</p>
				{:else if pushNotificationStore.syncError}
					<p
						class="mt-3 rounded-[12px] border border-[var(--color-danger)] bg-[var(--color-danger-soft)] px-3 py-2 text-sm text-[var(--color-danger)]"
					>
						{pushNotificationStore.syncError}
					</p>
				{/if}

				{#if pushNotificationStore.token}
					<div
						class="mt-3 rounded-[12px] border border-[var(--color-border-strong)] bg-[var(--color-surface)] p-3"
					>
						<p class="text-xs font-semibold text-[var(--color-text-muted)] uppercase">FCM token</p>
						<p class="mt-2 text-xs break-all text-[var(--color-text)]">
							{pushNotificationStore.token}
						</p>
					</div>
				{/if}

				{#if pushNotificationStore.lastMessage}
					<p class="mt-3 text-sm text-[var(--color-text-secondary)]">
						{pushNotificationStore.lastMessage}
					</p>
				{/if}
			</div>
		</div>
	</div>
</section>
