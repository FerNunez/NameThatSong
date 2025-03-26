// Custom JavaScript for the application

// Initialize album dropdown functionality
document.addEventListener('DOMContentLoaded', function() {
    // Setup album dropdown toggle functionality
    window.setSearchValue = function(name, id) {
        document.querySelector('input[name="search"]').value = name;
        document.querySelector('input[name="artist-id"]').value = id;
        document.getElementById('search-results').innerHTML = '';
    };
    
    window.scrollAlbumsLeft = function() {
        const scrollWrapper = document.querySelector('.album-scroll-wrapper');
        if (scrollWrapper) {
            scrollWrapper.scrollBy({ left: -300, behavior: 'smooth' });
        }
    };
    
    window.scrollAlbumsRight = function() {
        const scrollWrapper = document.querySelector('.album-scroll-wrapper');
        if (scrollWrapper) {
            scrollWrapper.scrollBy({ left: 300, behavior: 'smooth' });
        }
    };
    
    window.toggleAlbumDropdown = function() {
        const dropdown = document.querySelector('.album-dropdown-inner');
        const toggleButton = document.getElementById('toggle-album-button');
        
        if (!dropdown || !toggleButton) return;
        
        if (dropdown.style.display === 'none') {
            dropdown.style.display = 'block';
            toggleButton.textContent = 'Hide Albums';
        } else {
            dropdown.style.display = 'none';
            toggleButton.textContent = 'Show Albums';
            toggleButton.classList.remove('hidden');
        }
    };
    
    // Initialize the toggle button for the album dropdown
    const startButton = document.querySelector('.start-button');
    if (startButton) {
        startButton.addEventListener('click', function() {
            const toggleButton = document.getElementById('toggle-album-button');
            if (toggleButton) {
                toggleButton.classList.remove('hidden');
            }
        });
    }
});

// Add HTMX event listeners
document.addEventListener('htmx:afterSwap', function(event) {
    // Check if the album-dropdown-content was updated
    if (event.detail.target.id === 'album-dropdown-content') {
        const toggleButton = document.getElementById('toggle-album-button');
        if (toggleButton) {
            toggleButton.classList.add('hidden');
        }
    }
}); 