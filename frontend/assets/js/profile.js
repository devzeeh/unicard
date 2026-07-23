document.addEventListener("DOMContentLoaded", function () {
    console.log("Profile page script loaded.");

    const username = document.body.dataset.username;

    // --- Profile Edit Elements (Personal Info) ---
    const editProfileBtn = document.getElementById('edit-profile-btn');
    const cancelEditBtn = document.getElementById('cancel-edit-btn');
    const saveProfileBtn = document.getElementById('save-profile-btn');
    const profileActions = document.getElementById('profile-edit-actions');
    const profileView = document.getElementById('profile-details-view');
    const profileEditForm = document.getElementById('profile-details-edit');

    if (editProfileBtn && cancelEditBtn && profileActions && profileView && profileEditForm && saveProfileBtn) {
        editProfileBtn.addEventListener('click', () => {
            document.getElementById('full_name').value = document.getElementById('profile-view-name').innerText.trim();
            document.getElementById('username').value = document.getElementById('profile-view-username').innerText.trim();

            profileView.classList.add('hidden');
            profileEditForm.classList.remove('hidden');
            editProfileBtn.classList.add('hidden');
            profileActions.classList.remove('hidden');
        });

        cancelEditBtn.addEventListener('click', (e) => {
            e.preventDefault();
            profileEditForm.classList.add('hidden');
            profileView.classList.remove('hidden');
            profileActions.classList.add('hidden');
            editProfileBtn.classList.remove('hidden');
        });

        saveProfileBtn.addEventListener('click', async (e) => {
            e.preventDefault();
            const newName = document.getElementById('full_name').value;
            const newUsername = document.getElementById('username').value;

            try {
                const response = await fetch(`/u/${username}/profile/edit`, {
                    method: 'PATCH',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        full_name: newName,
                        username: newUsername
                    })
                });

                const result = await response.json();

                if (!result.success) {
                    alert(result.message || 'Failed to update profile');
                    return;
                }

                if (newUsername && newUsername !== username) {
                    window.location.href = `/u/${encodeURIComponent(newUsername)}`;
                    return;
                }

                document.getElementById('profile-view-name').innerText = newName;
                document.getElementById('profile-view-username').innerText = newUsername;

                profileEditForm.classList.add('hidden');
                profileView.classList.remove('hidden');
                profileActions.classList.add('hidden');
                editProfileBtn.classList.remove('hidden');

            } catch (err) {
                console.error('Profile update error:', err);
                alert('Network error, please try again.');
            }
        });
    }

    // --- Contact Edit Elements (Contact Info) ---
    const editContactBtn = document.getElementById('edit-contact-btn');
    const cancelContactBtn = document.getElementById('cancel-contact-edit-btn');
    const saveContactBtn = document.getElementById('save-contact-btn');
    const contactActions = document.getElementById('contact-edit-actions');
    const contactView = document.getElementById('contact-details-view');
    const contactEditForm = document.getElementById('contact-details-edit');

    if (editContactBtn && cancelContactBtn && contactActions && contactView && contactEditForm && saveContactBtn) {
        editContactBtn.addEventListener('click', () => {
            document.getElementById('email').value = document.getElementById('profile-view-email').innerText.trim();
            document.getElementById('phone').value = document.getElementById('profile-view-phone').innerText.trim();

            contactView.classList.add('hidden');
            contactEditForm.classList.remove('hidden');
            editContactBtn.classList.add('hidden');
            contactActions.classList.remove('hidden');
        });

        cancelContactBtn.addEventListener('click', (e) => {
            e.preventDefault();
            contactEditForm.classList.add('hidden');
            contactView.classList.remove('hidden');
            contactActions.classList.add('hidden');
            editContactBtn.classList.remove('hidden');
        });

        saveContactBtn.addEventListener('click', async (e) => {
            e.preventDefault();
            const newEmail = document.getElementById('email').value;
            const newPhone = document.getElementById('phone').value;

            try {
                const response = await fetch(`/u/${username}/profile/edit`, {
                    method: 'PATCH',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        email: newEmail,
                        phone_number: newPhone
                    })
                });

                const result = await response.json();

                if (!result.success) {
                    alert(result.message || 'Failed to update contact info');
                    return;
                }
                
                alert(result.message); // Will show success or "check email to verify"

                if (newEmail !== document.getElementById('profile-view-email').innerText.trim()) {
                    // Force refresh to fetch dashboard data and show pending email
                    window.location.reload();
                    return;
                }

                document.getElementById('profile-view-phone').innerText = newPhone;

                contactEditForm.classList.add('hidden');
                contactView.classList.remove('hidden');
                contactActions.classList.add('hidden');
                editContactBtn.classList.remove('hidden');

            } catch (err) {
                console.error('Contact info update error:', err);
                alert('Network error, please try again.');
            }
        });
    }
});