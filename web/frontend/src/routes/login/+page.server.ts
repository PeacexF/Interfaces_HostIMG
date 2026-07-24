import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
import { bffFetch, forwardSessionCookie } from '$lib/server/bff';

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const form = await request.formData();
		const email = form.get('email');
		const password = form.get('password');

		const res = await bffFetch('/api/login', '', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ email, password })
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({ error: 'login failed' }));
			return fail(res.status, { error: body.error as string });
		}

		forwardSessionCookie(res, cookies);
		redirect(303, '/');
	}
};