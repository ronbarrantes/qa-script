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
        
        // Ignore file drops when settings modal is active (could be internal drag-drop)
        const settingsModal = document.getElementById('settings-modal');
        if (settingsModal && settingsModal.classList.contains('active')) {
            return;
        }
        
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
        
        // Initialize drag-drop on the groups list
        initGroupsListDragDrop();
        
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
    groupItem.draggable = true;
    groupItem.innerHTML = `
        <div class="group-row">
            <div class="drag-handle" title="Drag to reorder">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="9" cy="6" r="1.5"></circle>
                    <circle cx="15" cy="6" r="1.5"></circle>
                    <circle cx="9" cy="12" r="1.5"></circle>
                    <circle cx="15" cy="12" r="1.5"></circle>
                    <circle cx="9" cy="18" r="1.5"></circle>
                    <circle cx="15" cy="18" r="1.5"></circle>
                </svg>
            </div>
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
    
    // Add drag event listeners (only dragstart and dragend on items)
    groupItem.addEventListener('dragstart', handleDragStart);
    groupItem.addEventListener('dragend', handleDragEnd);
    
    groupsList.appendChild(groupItem);
}

// Initialize drag-drop on the groups list container
function initGroupsListDragDrop() {
    const groupsList = document.getElementById('groups-list');
    if (!groupsList || groupsList.dataset.dragInitialized) return;
    
    groupsList.dataset.dragInitialized = 'true';
    
    groupsList.addEventListener('dragover', handleListDragOver);
    groupsList.addEventListener('drop', handleListDrop);
    
    // Prevent the modal from letting drag events bubble to the main window
    const modal = document.getElementById('settings-modal');
    if (modal && !modal.dataset.dragInitialized) {
        modal.dataset.dragInitialized = 'true';
        
        // Prevent file drops on the modal from reaching the main window
        modal.addEventListener('dragover', (e) => {
            if (draggedItem) {
                e.preventDefault();
                e.stopPropagation();
            }
        });
        modal.addEventListener('drop', (e) => {
            if (draggedItem) {
                e.preventDefault();
                e.stopPropagation();
            }
        });
        modal.addEventListener('dragenter', (e) => {
            if (draggedItem) {
                e.stopPropagation();
            }
        });
        modal.addEventListener('dragleave', (e) => {
            if (draggedItem) {
                e.stopPropagation();
            }
        });
    }
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

// =====================
// Drag and Drop for Groups
// =====================

let draggedItem = null;
let currentDropTarget = null;
let dropPosition = null; // 'before' or 'after'

function handleDragStart(e) {
    // Stop propagation to prevent Wails file drop handler from catching this
    e.stopPropagation();
    
    draggedItem = this;
    this.classList.add('dragging');
    
    // Set drag data - use a custom type to distinguish from file drops
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('application/x-group-drag', 'true');
    
    // Delay to allow drag image capture
    setTimeout(() => {
        const groupsList = document.getElementById('groups-list');
        groupsList.classList.add('dragging-active');
    }, 10);
}

function handleDragEnd(e) {
    e.stopPropagation();
    
    this.classList.remove('dragging');
    
    const groupsList = document.getElementById('groups-list');
    groupsList.classList.remove('dragging-active');
    
    // Remove all drag-over classes
    document.querySelectorAll('.group-item').forEach(item => {
        item.classList.remove('drag-over', 'drag-over-top', 'drag-over-bottom');
    });
    
    draggedItem = null;
    currentDropTarget = null;
    dropPosition = null;
}

function handleListDragOver(e) {
    // Only handle our custom group drag, not file drops
    if (!draggedItem) return;
    
    e.preventDefault();
    e.stopPropagation();
    e.dataTransfer.dropEffect = 'move';
    
    const groupsList = document.getElementById('groups-list');
    const items = [...groupsList.querySelectorAll('.group-item:not(.dragging)')];
    
    // Clear previous indicators
    items.forEach(item => {
        item.classList.remove('drag-over-top', 'drag-over-bottom');
    });
    
    // Find the item we're hovering over
    let targetItem = null;
    let insertBefore = true;
    
    for (const item of items) {
        const rect = item.getBoundingClientRect();
        const midY = rect.top + rect.height / 2;
        
        if (e.clientY < midY) {
            targetItem = item;
            insertBefore = true;
            break;
        } else if (e.clientY >= rect.top && e.clientY <= rect.bottom) {
            targetItem = item;
            insertBefore = false;
        }
    }
    
    // If we're below all items, target the last item
    if (!targetItem && items.length > 0) {
        targetItem = items[items.length - 1];
        insertBefore = false;
    }
    
    if (targetItem) {
        currentDropTarget = targetItem;
        dropPosition = insertBefore ? 'before' : 'after';
        
        if (insertBefore) {
            targetItem.classList.add('drag-over-top');
        } else {
            targetItem.classList.add('drag-over-bottom');
        }
    }
}

function handleListDrop(e) {
    // Only handle our custom group drag
    if (!draggedItem || !currentDropTarget) return;
    
    e.preventDefault();
    e.stopPropagation();
    
    const groupsList = document.getElementById('groups-list');
    
    // Perform the move
    if (dropPosition === 'before') {
        groupsList.insertBefore(draggedItem, currentDropTarget);
    } else {
        // Insert after
        const nextSibling = currentDropTarget.nextSibling;
        if (nextSibling) {
            groupsList.insertBefore(draggedItem, nextSibling);
        } else {
            groupsList.appendChild(draggedItem);
        }
    }
    
    // Clean up
    document.querySelectorAll('.group-item').forEach(item => {
        item.classList.remove('drag-over', 'drag-over-top', 'drag-over-bottom');
    });
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
