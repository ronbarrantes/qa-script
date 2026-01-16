// State
let locationsFile = '';
let prioritiesFile = '';
let lastOutputPath = '';

// Settings state
let savedSettings = null; // Last saved settings (for reset functionality)

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    // Make drop zones clickable
    document.getElementById('locations-zone').addEventListener('click', () => {
        selectFile('locations');
    });
    document.getElementById('priorities-zone').addEventListener('click', () => {
        selectFile('priorities');
    });

    // Handle Wails native file drop events (required for Windows)
    // The event provides x, y coordinates and files array
    window.runtime?.EventsOn('wails:file-drop', (x, y, files) => {
        if (!files || files.length === 0) return;
        
        const file = files[0];
        const locationsZone = document.getElementById('locations-zone');
        const prioritiesZone = document.getElementById('priorities-zone');
        
        // Get element at drop coordinates to determine which zone
        const dropTarget = document.elementFromPoint(x, y);
        
        // Check if drop was on locations zone
        if (locationsZone.contains(dropTarget) || dropTarget === locationsZone) {
            if (file.toLowerCase().endsWith('.csv')) {
                setFile('locations', file);
            } else {
                showStatus('Please drop a CSV file for locations', 'error');
                showZoneError('locations');
            }
            return;
        }
        
        // Check if drop was on priorities zone
        if (prioritiesZone.contains(dropTarget) || dropTarget === prioritiesZone) {
            if (file.toLowerCase().endsWith('.xlsx')) {
                setFile('priorities', file);
            } else {
                showStatus('Please drop an XLSX file for priorities', 'error');
                showZoneError('priorities');
            }
            return;
        }
        
        // File dropped outside specific zones - auto-detect by extension
        if (file.toLowerCase().endsWith('.csv')) {
            setFile('locations', file);
        } else if (file.toLowerCase().endsWith('.xlsx')) {
            setFile('priorities', file);
        } else {
            showStatus('Unsupported file type. Please use CSV or XLSX files.', 'error');
        }
    });
});

// Drag and drop handlers
function handleDragOver(event) {
    event.preventDefault();
    event.currentTarget.classList.add('drag-over');
}

function handleDragLeave(event) {
    event.currentTarget.classList.remove('drag-over');
}

function handleDrop(event, type) {
    event.preventDefault();
    event.currentTarget.classList.remove('drag-over');
    // On Windows, native file drop is handled by Wails via wails:file-drop event
    // The webview's dataTransfer doesn't contain proper file paths on Windows
}

// File selection via dialog
async function selectFile(type) {
    try {
        let path;
        if (type === 'locations') {
            path = await window.go.main.App.SelectLocationsFile();
        } else {
            path = await window.go.main.App.SelectPrioritiesFile();
        }
        
        if (path) {
            updateFileDisplay(type, path);
        }
    } catch (err) {
        showStatus('Error: ' + err, 'error');
        showZoneError(type);
    }
}

// Set file from drag and drop
async function setFile(type, path) {
    try {
        if (type === 'locations') {
            await window.go.main.App.SetLocationsFile(path);
        } else {
            await window.go.main.App.SetPrioritiesFile(path);
        }
        updateFileDisplay(type, path);
    } catch (err) {
        showStatus('Error: ' + err, 'error');
        showZoneError(type);
    }
}

// Update UI to show selected file
function updateFileDisplay(type, path) {
    const zone = document.getElementById(type + '-zone');
    const fileNameEl = document.getElementById(type + '-file');
    
    // Extract just the filename
    const fileName = path.split('/').pop().split('\\').pop();
    fileNameEl.textContent = fileName;
    zone.classList.add('has-file');
    zone.classList.remove('error');
    
    if (type === 'locations') {
        locationsFile = path;
    } else {
        prioritiesFile = path;
    }
    
    updateOkButton();
    clearStatus();
}

// Show error on drop zone
function showZoneError(type) {
    const zone = document.getElementById(type + '-zone');
    zone.classList.add('error');
    setTimeout(() => zone.classList.remove('error'), 400);
}

// Update OK button state
function updateOkButton() {
    const okBtn = document.getElementById('ok-btn');
    okBtn.disabled = !(locationsFile && prioritiesFile);
}

// Reset files
async function resetFiles() {
    await window.go.main.App.Reset();
    
    locationsFile = '';
    prioritiesFile = '';
    
    document.getElementById('locations-file').textContent = '';
    document.getElementById('priorities-file').textContent = '';
    document.getElementById('locations-zone').classList.remove('has-file', 'error');
    document.getElementById('priorities-zone').classList.remove('has-file', 'error');
    
    updateOkButton();
    clearStatus();
}

// Process files
async function processFiles() {
    const okBtn = document.getElementById('ok-btn');
    okBtn.disabled = true;
    showStatus('Processing...', 'processing');
    
    try {
        const outputPath = await window.go.main.App.Process();
        lastOutputPath = outputPath;
        // Get shortened path for display (e.g., ~/Downloads/file.xlsx)
        const displayPath = await window.go.main.App.ShortenPath(outputPath);
        showStatusWithOpen('✓ ' + displayPath, outputPath);
    } catch (err) {
        showStatus('Error: ' + err, 'error');
    } finally {
        okBtn.disabled = false;
    }
}

// Open the output file
async function openOutputFile() {
    if (!lastOutputPath) return;
    try {
        await window.go.main.App.OpenFile(lastOutputPath);
    } catch (err) {
        showStatus('Error opening file: ' + err, 'error');
    }
}

// Status display helpers
function showStatus(message, type) {
    const status = document.getElementById('status');
    const statusMessage = document.getElementById('status-message');
    const openBtn = document.getElementById('open-file-btn');
    
    statusMessage.textContent = message;
    status.className = 'status ' + type;
    openBtn.style.display = 'none';
}

function showStatusWithOpen(message, fullPath) {
    const status = document.getElementById('status');
    const statusMessage = document.getElementById('status-message');
    const openBtn = document.getElementById('open-file-btn');
    
    statusMessage.textContent = message;
    statusMessage.title = fullPath; // Show full path on hover
    status.className = 'status success';
    openBtn.style.display = 'inline-block';
}

function clearStatus() {
    const status = document.getElementById('status');
    const statusMessage = document.getElementById('status-message');
    const openBtn = document.getElementById('open-file-btn');
    
    statusMessage.textContent = '';
    statusMessage.title = '';
    status.className = 'status';
    openBtn.style.display = 'none';
    lastOutputPath = '';
}

// =====================
// Settings Modal
// =====================

// Open settings modal
async function openSettings() {
    const modal = document.getElementById('settings-modal');
    
    try {
        // Load current settings from backend
        const config = await window.go.main.App.GetRulesConfig();
        savedSettings = JSON.parse(JSON.stringify(config)); // Deep clone for reset
        
        // Populate the UI
        populateSettingsUI(config);
        
        // Show modal
        modal.classList.add('active');
    } catch (err) {
        showStatus('Error loading settings: ' + err, 'error');
    }
}

// Close settings modal
function closeSettings() {
    const modal = document.getElementById('settings-modal');
    modal.classList.remove('active');
}

// Populate settings UI with config data
function populateSettingsUI(config) {
    const groupsList = document.getElementById('groups-list');
    groupsList.innerHTML = '';
    
    // Add groups
    config.groups.forEach((group, index) => {
        addGroupToUI(group.title, group.values, index);
    });
    
    // Set numeric fields
    document.getElementById('max-rows').value = config.maxRows || 20;
    document.getElementById('column-gap').value = config.columnGap ?? 1;
}

// Add a group item to the UI
function addGroupToUI(title = '', values = '', index = null) {
    const groupsList = document.getElementById('groups-list');
    
    const groupItem = document.createElement('div');
    groupItem.className = 'group-item';
    groupItem.innerHTML = `
        <div class="group-row">
            <input type="text" class="group-title-input" placeholder="Group Title" value="${escapeHtml(title)}">
            <input type="text" class="group-values-input" placeholder="a, b, c, ..." value="${escapeHtml(values)}">
            <button class="remove-group-btn" onclick="removeGroup(this)" title="Remove group">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
            </button>
        </div>
    `;
    
    groupsList.appendChild(groupItem);
}

// Add a new empty group
function addGroup() {
    addGroupToUI('', '');
    
    // Focus the new title input
    const groupsList = document.getElementById('groups-list');
    const lastGroup = groupsList.lastElementChild;
    if (lastGroup) {
        const titleInput = lastGroup.querySelector('.group-title-input');
        titleInput.focus();
    }
}

// Remove a group
function removeGroup(button) {
    const groupItem = button.closest('.group-item');
    groupItem.style.opacity = '0';
    groupItem.style.transform = 'translateX(-10px)';
    setTimeout(() => {
        groupItem.remove();
    }, 150);
}

// Get current settings from UI
function getSettingsFromUI() {
    const groupsList = document.getElementById('groups-list');
    const groupItems = groupsList.querySelectorAll('.group-item');
    
    const groups = [];
    groupItems.forEach(item => {
        const title = item.querySelector('.group-title-input').value.trim();
        const values = item.querySelector('.group-values-input').value.trim();
        
        // Only include groups with a title
        if (title) {
            groups.push({ title, values });
        }
    });
    
    return {
        groups: groups,
        maxRows: parseInt(document.getElementById('max-rows').value) || 20,
        columnGap: parseInt(document.getElementById('column-gap').value) || 0
    };
}

// Reset settings to last saved state
function resetSettings() {
    if (savedSettings) {
        populateSettingsUI(savedSettings);
    }
}

// Reset to default settings
async function resetToDefaults() {
    try {
        const defaultConfig = await window.go.main.App.GetDefaultRulesConfig();
        populateSettingsUI(defaultConfig);
    } catch (err) {
        showStatus('Error loading defaults: ' + err, 'error');
    }
}

// Save settings
async function saveSettings() {
    const config = getSettingsFromUI();
    
    // Validate
    if (config.groups.length === 0) {
        showStatus('At least one group is required', 'error');
        return;
    }
    
    if (config.maxRows < 1) {
        showStatus('Max rows must be at least 1', 'error');
        return;
    }
    
    try {
        await window.go.main.App.SaveRulesConfig(config);
        savedSettings = JSON.parse(JSON.stringify(config)); // Update saved state
        closeSettings();
        showStatus('Settings saved', 'success');
        setTimeout(clearStatus, 2000);
    } catch (err) {
        showStatus('Error saving settings: ' + err, 'error');
    }
}

// Escape HTML for safe insertion
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Close modal on outside click
document.addEventListener('click', (e) => {
    const modal = document.getElementById('settings-modal');
    if (e.target === modal) {
        closeSettings();
    }
});

// Close modal on Escape key
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        const modal = document.getElementById('settings-modal');
        if (modal.classList.contains('active')) {
            closeSettings();
        }
    }
});
