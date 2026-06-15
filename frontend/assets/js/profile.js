document.addEventListener("DOMContentLoaded", function () {
    console.log("Profile page script loaded.");

    const username = document.body.dataset.username;

    // --- Profile Edit Elements ---
    const editProfileBtn = document.getElementById('edit-profile-btn');
    const cancelEditBtn = document.getElementById('cancel-edit-btn');
    const saveProfileBtn = document.getElementById('save-profile-btn');
    const profileActions = document.getElementById('profile-edit-actions');
    const profileView = document.getElementById('profile-details-view');
    const profileEditForm = document.getElementById('profile-details-edit');

    if (editProfileBtn && cancelEditBtn && profileActions && profileView && profileEditForm && saveProfileBtn) {
        editProfileBtn.addEventListener('click', () => {
            // Pre-fill edit form with current values
            document.getElementById('full_name').value = document.getElementById('profile-view-name').innerText.trim();
            document.getElementById('email').value = document.getElementById('profile-view-email').innerText.trim();
            document.getElementById('phone').value = document.getElementById('profile-view-phone').innerText.trim();
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
            const newEmail = document.getElementById('email').value;
            const newPhone = document.getElementById('phone').value;
            const newUsername = document.getElementById('username').value;

            try {
                const response = await fetch(`/u/${username}/profile/edit`, {
                    method: 'PATCH',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        full_name: newName,
                        email: newEmail,
                        phone_number: newPhone,
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

                const nameSpan = document.getElementById('profile-view-name');
                const emailSpan = document.getElementById('profile-view-email');
                const phoneSpan = document.getElementById('profile-view-phone');
                const usernameSpan = document.getElementById('profile-view-username');

                if (nameSpan) nameSpan.innerText = newName;
                if (emailSpan) emailSpan.innerText = newEmail;
                if (phoneSpan) phoneSpan.innerText = newPhone;
                if (usernameSpan) usernameSpan.innerText = newUsername;

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
});