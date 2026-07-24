<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<h1>Your files</h1>

<form method="POST" action="?/upload" enctype="multipart/form-data" use:enhance>
	<input type="file" name="file" required />
	<button type="submit">Upload</button>
</form>

{#if form?.error}
	<p>{form.error}</p>
{/if}

<ul>
	{#each data.files as file (file.id)}
		<li>
			{#if file.mime_type.startsWith('image/')}
				<img
					src={resolve('/api/files/[id]/thumbnail', { id: file.id })}
					alt={file.name}
					width="128"
				/>
			{/if}
			<a href={resolve('/api/files/[id]', { id: file.id })}>{file.name}</a>
			({file.size} bytes)
			<form method="POST" action="?/delete" use:enhance style="display:inline">
				<input type="hidden" name="id" value={file.id} />
				<button type="submit">Delete</button>
			</form>
		</li>
	{/each}
</ul>