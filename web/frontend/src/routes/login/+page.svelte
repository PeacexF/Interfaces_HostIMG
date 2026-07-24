<script lang="ts">
	import type { ActionData } from './$types';
	let { form }: { form: ActionData } = $props();
</script>

<h1>Log in</h1>

{#if form?.error}
	<p>{form.error}</p>
{/if}

<form method="POST">
	<input type="email" name="email" placeholder="Email" required />
	<input type="password" name="password" placeholder="Password" required />
	<button type="submit">Log in</button>
</form>