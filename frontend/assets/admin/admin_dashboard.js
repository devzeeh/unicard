document.addEventListener("DOMContentLoaded", function () {
    fetch('/v1/admin/dashboard-data', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({})
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                document.getElementById('grossRevenue').textContent = '₱' + data.data.grossRevenue.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
                document.getElementById('netRevenue').textContent = '₱' + data.data.netRevenue.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
                document.getElementById('totalUsers').textContent = data.data.totalUsers.toLocaleString();
                document.getElementById('totalCards').textContent = data.data.totalCards.toLocaleString();
                document.getElementById('activeMerchants').textContent = data.data.activeMerchants.toLocaleString();
                document.getElementById('activeTerminals').textContent = data.data.activeTerminals.toLocaleString();
            } else {
                console.error("Failed to load dashboard data:", data.message);
            }
        })
        .catch(error => console.error("Error fetching dashboard data:", error));
});