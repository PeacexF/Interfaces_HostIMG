<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import { resolve } from '$app/paths';
	import type { Snippet } from 'svelte';
	import type { LayoutData } from './$types';

	let { children, data }: { children: Snippet; data: LayoutData } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<nav>
	<a href={resolve('/')}>Gallery</a>
	{#if data.accountId}
		<a href={resolve('/link')}>Link Telegram</a>
		<form method="POST" action={resolve('/logout')}>
			<button type="submit">Log out</button>
		</form>
	{:else}
		<a href={resolve('/login')}>Log in</a>
		<a href={resolve('/signup')}>Sign up</a>
	{/if}
</nav>

{@render children()}