<script lang="ts">
	let {
		value = $bindable(),
		autocomplete = 'current-password',
		placeholder = '',
		required = false,
		minlength,
		ariaLabel = 'Mật khẩu',
		inputClass = 'h-11'
	}: {
		value: string;
		autocomplete?: 'current-password' | 'new-password' | 'off';
		placeholder?: string;
		required?: boolean;
		minlength?: number;
		ariaLabel?: string;
		inputClass?: string;
	} = $props();

	let visible = $state(false);
</script>

<div class="relative">
	<input
		bind:value
		type={visible ? 'text' : 'password'}
		{autocomplete}
		{placeholder}
		{required}
		{minlength}
		aria-label={ariaLabel}
		class={[inputClass, 'w-full rounded-md border-[var(--color-border-strong)] pr-11']}
	/>
	<button
		type="button"
		class="absolute inset-y-0 right-0 grid w-11 place-items-center text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
		aria-label={visible ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'}
		aria-pressed={visible}
		title={visible ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'}
		onclick={() => (visible = !visible)}
	>
		<span
			class={visible ? 'icon-[lucide--eye-off] h-5 w-5' : 'icon-[lucide--eye] h-5 w-5'}
			aria-hidden="true"
		></span>
	</button>
</div>
