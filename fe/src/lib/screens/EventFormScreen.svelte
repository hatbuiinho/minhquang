<script lang="ts">
	import { onMount } from 'svelte';

	import { eventStore, reminderPresetOptions } from '$lib/events/event-store.svelte';
	import { router } from '$lib/navigation/router.svelte';

	let { mode, eventId }: { mode: 'create' | 'edit'; eventId?: string } = $props();
	let titleInput: HTMLInputElement | undefined;
	const inputClass =
		'mt-1 h-11 w-full rounded-[12px] border-[var(--color-border-strong)] bg-[var(--color-surface)] text-base text-[var(--color-text)] focus:border-[var(--color-primary)] focus:ring-[var(--color-primary)]';
	const textareaClass =
		'mt-1 min-h-32 w-full rounded-[12px] border-[var(--color-border-strong)] bg-[var(--color-surface)] text-base text-[var(--color-text)] focus:border-[var(--color-primary)] focus:ring-[var(--color-primary)]';
	const fieldsetClass =
		'rounded-[16px] border border-[var(--color-border)] bg-[var(--color-surface)] p-3 shadow-[var(--shadow-soft)]';
	const optionClass =
		'flex min-h-11 items-center gap-2 rounded-[12px] bg-[var(--color-bg)] px-3 text-[var(--color-text)]';
	const checkboxClass =
		'rounded border-[var(--color-border-strong)] text-[var(--color-primary)] focus:ring-[var(--color-primary)]';
	const radioClass =
		'border-[var(--color-border-strong)] text-[var(--color-primary)] focus:ring-[var(--color-primary)]';

	onMount(() => {
		void eventStore.loadAudienceOptions();
		if (mode === 'create') {
			requestAnimationFrame(() => titleInput?.focus());
		}
	});

	async function save() {
		const event = await eventStore.saveForm();
		if (event) {
			router.replace(mode === 'create' ? '/events' : `/events/${encodeURIComponent(event.id)}`);
		}
	}

	function cancel() {
		if (mode === 'edit' && eventId) {
			router.replace(`/events/${encodeURIComponent(eventId)}`);
			return;
		}

		router.replace('/events');
	}
</script>

<section class="flex h-full flex-col">
	<form
		id="event-form"
		class="flex-1 space-y-4 overflow-y-auto px-4 py-4 pb-28"
		onsubmit={(submitEvent) => {
			submitEvent.preventDefault();
			void save();
		}}
	>
		<label class="block">
			<span class="text-sm font-medium text-[var(--color-text)]">Tiêu đề</span>
			<input
				bind:this={titleInput}
				class={inputClass}
				bind:value={eventStore.form.title}
				required
			/>
		</label>

		<label class="block">
			<span class="text-sm font-medium text-[var(--color-text)]">Thời gian diễn ra</span>
			<input
				type="datetime-local"
				class={inputClass}
				bind:value={eventStore.form.starts_at}
				required
			/>
		</label>

		<label class="block">
			<span class="text-sm font-medium text-[var(--color-text)]">Múi giờ</span>
			<input class={inputClass} bind:value={eventStore.form.timezone} />
		</label>

		<label class="block">
			<span class="text-sm font-medium text-[var(--color-text)]">Trạng thái</span>
			<select class={inputClass} bind:value={eventStore.form.status}>
				<option value="active">Đang hoạt động</option>
				<option value="archived">Đã lưu trữ</option>
				<option value="cancelled">Đã huỷ</option>
			</select>
		</label>

		<label class="block">
			<span class="text-sm font-medium text-[var(--color-text)]">Mô tả</span>
			<textarea class={textareaClass} bind:value={eventStore.form.description}></textarea>
		</label>

		<fieldset class={fieldsetClass}>
			<legend class="px-1 text-sm font-medium text-[var(--color-text)]">
				<span class="inline-flex items-center gap-2">
					<span class="icon-[lucide--users] h-4 w-4" aria-hidden="true"></span>
					Người nhận
				</span>
			</legend>
			<div class="mt-2 grid grid-cols-2 gap-2">
				<label class={optionClass}>
					<input
						type="radio"
						class={radioClass}
						bind:group={eventStore.form.audience_type}
						value="self"
					/>
					<span class="text-sm">Chỉ mình tôi</span>
				</label>
				<label class={optionClass}>
					<input
						type="radio"
						class={radioClass}
						bind:group={eventStore.form.audience_type}
						value="selected_users"
					/>
					<span class="text-sm">Chọn từng người</span>
				</label>
				<label class={optionClass}>
					<input
						type="radio"
						class={radioClass}
						bind:group={eventStore.form.audience_type}
						value="selected_groups"
					/>
					<span class="text-sm">Tất cả theo group</span>
				</label>
				<label class={optionClass}>
					<input
						type="radio"
						class={radioClass}
						bind:group={eventStore.form.audience_type}
						value="all_users"
					/>
					<span class="text-sm">Tất cả user</span>
				</label>
			</div>

			{#if eventStore.isLoadingAudience}
				<p class="mt-3 text-sm text-[var(--color-text-secondary)]">
					Đang tải danh sách người nhận...
				</p>
			{:else if eventStore.form.audience_type === 'selected_users'}
				<div class="mt-3 space-y-2">
					{#each eventStore.users as user (user.id)}
						<label
							class="flex min-h-10 items-center gap-3 rounded-[12px] bg-[var(--color-bg)] px-3"
						>
							<input
								type="checkbox"
								class={checkboxClass}
								checked={eventStore.form.recipient_user_ids.includes(user.id)}
								onchange={() => eventStore.toggleRecipientUser(user.id)}
							/>
							<span class="min-w-0 flex-1 text-sm">
								<span class="block truncate">{user.name}</span>
								{#if user.email}
									<span class="block truncate text-xs text-[var(--color-text-secondary)]"
										>{user.email}</span
									>
								{/if}
							</span>
						</label>
					{:else}
						<p class="text-sm text-[var(--color-text-secondary)]">Chưa có user để chọn.</p>
					{/each}
				</div>
			{:else if eventStore.form.audience_type === 'selected_groups'}
				<div class="mt-3 space-y-2">
					{#each eventStore.groups as group (group.id)}
						<label
							class="flex min-h-10 items-center gap-3 rounded-[12px] bg-[var(--color-bg)] px-3"
						>
							<input
								type="checkbox"
								class={checkboxClass}
								checked={eventStore.form.recipient_group_ids.includes(group.id)}
								onchange={() => eventStore.toggleRecipientGroup(group.id)}
							/>
							<span class="min-w-0 flex-1 text-sm">
								<span class="block truncate">{group.name}</span>
								{#if group.description}
									<span class="block truncate text-xs text-[var(--color-text-secondary)]"
										>{group.description}</span
									>
								{/if}
							</span>
						</label>
					{:else}
						<p class="text-sm text-[var(--color-text-secondary)]">Chưa có group để lọc.</p>
					{/each}
				</div>
			{/if}
		</fieldset>

		<fieldset class={fieldsetClass}>
			<legend class="px-1 text-sm font-medium text-[var(--color-text)]">
				<span class="inline-flex items-center gap-2">
					<span class="icon-[lucide--clock] h-4 w-4" aria-hidden="true"></span>
					Nhắc trước
				</span>
			</legend>
			<div class="mt-2 space-y-2">
				{#each reminderPresetOptions as option (option.offsetMinutes)}
					<label class="flex min-h-11 items-center gap-3 rounded-[12px] bg-[var(--color-bg)] px-3">
						<input
							type="checkbox"
							class={checkboxClass}
							checked={eventStore.form.reminder_offsets.includes(option.offsetMinutes)}
							onchange={() => eventStore.toggleReminderOffset(option.offsetMinutes)}
						/>
						<span class="text-sm">{option.label}</span>
					</label>
				{/each}
			</div>

			<div class="mt-4">
				<p class="text-sm font-medium text-[var(--color-text)]">Mức thông báo</p>
				<div class="mt-2 grid grid-cols-2 gap-2">
					<label class={optionClass}>
						<input
							type="radio"
							class={radioClass}
							bind:group={eventStore.form.reminder_importance}
							value="normal"
						/>
						<span class="text-sm">Bình thường</span>
					</label>
					<label class={optionClass}>
						<input
							type="radio"
							class={radioClass}
							bind:group={eventStore.form.reminder_importance}
							value="urgent"
						/>
						<span class="text-sm">Quan trọng</span>
					</label>
				</div>
			</div>
		</fieldset>
	</form>

	<div
		class="border-t border-[var(--color-border)] bg-[rgb(255_255_255_/_0.94)] px-4 py-3 pb-[max(env(safe-area-inset-bottom),0.75rem)] backdrop-blur"
	>
		<div class="grid grid-cols-[0.8fr_1.2fr] gap-3">
			<button
				type="button"
				class="h-11 rounded-[12px] border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-sm font-semibold text-[var(--color-text)]"
				onclick={cancel}
			>
				Huỷ
			</button>
			<button
				type="submit"
				form="event-form"
				class="flex h-11 items-center justify-center gap-2 rounded-[12px] bg-[var(--color-primary)] text-sm font-semibold text-white shadow-[var(--shadow-soft)] disabled:opacity-60"
				disabled={eventStore.isSaving || !eventStore.form.title || !eventStore.form.starts_at}
				aria-label={eventStore.isSaving ? 'Đang lưu' : mode === 'edit' ? 'Lưu' : 'Tạo'}
			>
				<span class="icon-[lucide--check] h-5 w-5" aria-hidden="true"></span>
			</button>
		</div>
	</div>
</section>
