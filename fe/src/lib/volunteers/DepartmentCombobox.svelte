<script lang="ts">
	import { onMount } from 'svelte';
	import { listDepartmentSuggestions } from './api';

	let { value = $bindable() }: { value: string } = $props();
	let root = $state<HTMLDivElement>();
	let input = $state<HTMLInputElement>();
	let options = $state<string[]>([]);
	let open = $state(false);
	let loading = $state(false);
	let activeIndex = $state(-1);
	let debounceTimer: ReturnType<typeof setTimeout> | undefined;
	let requestController: AbortController | undefined;
	const debounceMs = 300;
	const listboxId = 'department-suggestions';

	function scheduleSuggestions() {
		if (debounceTimer) clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => void loadSuggestions(), debounceMs);
	}

	async function loadSuggestions() {
		requestController?.abort();
		const controller = new AbortController();
		requestController = controller;
		loading = true;
		try {
			options = await listDepartmentSuggestions(value, controller.signal);
			activeIndex = options.length > 0 ? 0 : -1;
		} catch (error) {
			if (!(error instanceof DOMException && error.name === 'AbortError')) options = [];
		} finally {
			if (requestController === controller) loading = false;
		}
	}

	function handleFocus() {
		open = true;
		void loadSuggestions();
	}

	function handleInput() {
		open = true;
		activeIndex = -1;
		scheduleSuggestions();
	}

	function selectOption(option: string) {
		value = option;
		open = false;
		activeIndex = -1;
		input?.focus();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'ArrowDown') {
			event.preventDefault();
			open = true;
			activeIndex = Math.min(activeIndex + 1, options.length - 1);
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			activeIndex = Math.max(activeIndex - 1, 0);
		} else if (event.key === 'Enter' && open && activeIndex >= 0) {
			event.preventDefault();
			selectOption(options[activeIndex]);
		} else if (event.key === 'Escape') {
			open = false;
			activeIndex = -1;
		} else if (event.key === 'Tab') {
			open = false;
		}
	}

	onMount(() => {
		function closeWhenClickingOutside(event: PointerEvent) {
			if (event.target instanceof Node && root && !root.contains(event.target)) open = false;
		}
		document.addEventListener('pointerdown', closeWhenClickingOutside);
		return () => {
			if (debounceTimer) clearTimeout(debounceTimer);
			requestController?.abort();
			document.removeEventListener('pointerdown', closeWhenClickingOutside);
		};
	});
</script>

<div class="relative" bind:this={root}>
	<input
		bind:this={input}
		bind:value
		maxlength="60"
		role="combobox"
		aria-label="Phân ban"
		aria-autocomplete="list"
		aria-expanded={open}
		aria-controls={listboxId}
		aria-activedescendant={activeIndex >= 0 ? `${listboxId}-${activeIndex}` : undefined}
		class="h-11 w-full rounded-md border-[var(--color-border-strong)] pr-9"
		onfocus={handleFocus}
		oninput={handleInput}
		onkeydown={handleKeydown}
	/>
	<span
		class={[
			'pointer-events-none absolute top-3 right-3 h-4 w-4 text-[var(--color-text-muted)]',
			loading ? 'icon-[lucide--loader-circle] animate-spin' : 'icon-[lucide--chevron-down]'
		]}
		aria-hidden="true"
	></span>

	{#if open && options.length > 0}
		<ul
			id={listboxId}
			role="listbox"
			class="absolute z-30 mt-1 max-h-52 w-full overflow-y-auto rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] p-1 shadow-[var(--shadow-popover)]"
		>
			{#each options as option, index (option)}
				<li id={`${listboxId}-${index}`} role="option" aria-selected={index === activeIndex}>
					<button
						type="button"
						class={[
							'w-full rounded px-3 py-2 text-left text-sm',
							index === activeIndex
								? 'bg-[var(--color-primary-soft)]'
								: 'hover:bg-[var(--color-surface-muted)]'
						]}
						onpointerdown={(event) => event.preventDefault()}
						onmouseenter={() => (activeIndex = index)}
						onclick={() => selectOption(option)}
					>
						{option}
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
