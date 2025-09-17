// ML Notes Wails App JavaScript

class MLNotesWailsApp {
    constructor() {
        console.log('MLNotesWailsApp constructor called');
        window.runtime.LogInfo('JavaScript: MLNotesWailsApp constructor called');

        this.currentNoteId = null;
        this.isNewNote = false;
        this.unsavedChanges = false;
        this.enhancedEditor = null;

        this.init();
    }

    async init() {
        try {
            // Check if app is initialized first
            const isInitialized = await this.checkInitialization();
            if (!isInitialized) {
                this.showInitializationPage();
                return;
            }

            this.setupEventListeners();
            this.setupTheme();
            await this.loadNotesAndTags();
            this.setupMarkdownPreview();
            this.initializeSlidingPanels();
        } catch (error) {
            console.error('Error during initialization:', error);
            // If initialization fails, still show the main app
            this.setupEventListeners();
            this.setupTheme();
            this.setupMarkdownPreview();
            this.initializeSlidingPanels();
        }
    }

    setupEventListeners() {
        // Theme toggle
        document.getElementById('theme-toggle').addEventListener('click', () => {
            this.toggleTheme();
        });

        // New note button
        document.getElementById('new-note-btn').addEventListener('click', () => {
            this.createNewNote();
        });

        // Create first note button (welcome screen)
        const createFirstNoteBtn = document.getElementById('create-first-note');
        if (createFirstNoteBtn) {
            createFirstNoteBtn.addEventListener('click', () => {
                this.createNewNote();
            });
        }

        // Editor controls
        const saveBtn = document.getElementById('save-note');
        if (saveBtn) {
            saveBtn.addEventListener('click', () => {
                this.saveCurrentNote();
            });
        }

        const deleteBtn = document.getElementById('delete-note');
        if (deleteBtn) {
            deleteBtn.addEventListener('click', () => {
                this.deleteCurrentNote();
            });
        }

        const autoTagBtn = document.getElementById('auto-tag-btn');
        if (autoTagBtn) {
            autoTagBtn.addEventListener('click', () => {
                this.autoTagNote();
            });
        }

        const analyzeBtn = document.getElementById('analyze-btn');
        if (analyzeBtn) {
            analyzeBtn.addEventListener('click', () => {
                this.analyzeNote();
            });
        }

        // Content change tracking
        const titleInput = document.getElementById('note-title');
        const noteContent = document.getElementById('note-content');
        const tagsInput = document.getElementById('note-tags');

        if (titleInput) {
            titleInput.addEventListener('input', () => {
                this.markUnsaved();
                this.updateDocumentTitle();
            });
        }

        // Initialize Enhanced Editor
        const noteContentTextarea = document.getElementById('note-content');
        const notePreview = document.getElementById('note-preview');

        if (noteContentTextarea && notePreview) {
            this.enhancedEditor = new EnhancedEditor(noteContentTextarea, notePreview);

            noteContentTextarea.addEventListener('input', () => {
                this.markUnsaved();
            });
        }

        if (tagsInput) {
            tagsInput.addEventListener('input', () => {
                this.markUnsaved();
            });
        }

        // Tag removal and note clicks (event delegation for performance)
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('tag-remove')) {
                const tag = e.target.closest('.tag');
                const tagValue = tag.dataset.tag;
                this.removeTag(tagValue);
            } else if (e.target.closest('.note-item')) {
                // Handle note item clicks
                const noteItem = e.target.closest('.note-item');
                const noteId = parseInt(noteItem.dataset.noteId);
                if (!isNaN(noteId)) {
                    this.loadNote(noteId);
                }
            }
        });

        // Search
        const searchInput = document.getElementById('search-input');
        const searchBtn = document.getElementById('search-btn');

        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.performSearch(e.target.value);
            });

            searchInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.performSearch(e.target.value);
                }
            });
        }

        if (searchBtn) {
            searchBtn.addEventListener('click', () => {
                const query = searchInput.value;
                this.performSearch(query);
            });
        }

        // Tag filter
        const tagFilter = document.getElementById('tag-filter');
        if (tagFilter) {
            tagFilter.addEventListener('change', (e) => {
                this.filterByTag(e.target.value);
            });
        }

        // Settings button
        const settingsBtn = document.getElementById('settings-btn');
        console.log('Settings button found:', !!settingsBtn);
        if (settingsBtn) {
            settingsBtn.addEventListener('click', () => {
                console.log('Settings button clicked!');
                window.runtime.LogInfo('JavaScript: Settings button clicked!');
                try {
                    console.log('About to call showSettingsPage...');
                    this.showSettingsPage();
                    console.log('showSettingsPage completed');
                } catch (error) {
                    console.error('Error opening settings page:', error);
                    window.runtime.LogError('JavaScript error: ' + error.message);
                    alert('Settings error: ' + error.message);
                }
            });
        } else {
            console.error('Settings button not found!');
        }

        // Settings page controls
        const closeSettingsBtn = document.getElementById('close-settings');
        if (closeSettingsBtn) {
            closeSettingsBtn.addEventListener('click', () => {
                this.hideSettingsPage();
            });
        }

        const settingsForm = document.getElementById('settings-form');
        if (settingsForm) {
            settingsForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.saveSettings();
            });
        }

        const testOllamaBtn = document.getElementById('test-ollama');
        if (testOllamaBtn) {
            testOllamaBtn.addEventListener('click', () => {
                this.testOllamaConnection();
            });
        }

        // Initialization page controls
        const initForm = document.getElementById('init-form');
        if (initForm) {
            initForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.initializeApp();
            });
        }

        // Modal controls
        this.setupModalControls();

        // Keyboard shortcuts
        this.setupKeyboardShortcuts();
        this.setupResponsiveActions();

        // Before unload warning
        window.addEventListener('beforeunload', (e) => {
            if (this.unsavedChanges) {
                e.preventDefault();
                e.returnValue = '';
            }
        });
    }

    setupModalControls() {
        // Close modals
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal-overlay') ||
                e.target.classList.contains('modal-close')) {
                this.closeModals();
            }
        });
    }

    setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            // Ctrl/Cmd + S = Save
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                this.saveCurrentNote();
            }

            // Ctrl/Cmd + N = New note
            if ((e.ctrlKey || e.metaKey) && e.key === 'n') {
                e.preventDefault();
                this.createNewNote();
            }

            // Ctrl/Cmd + / = Toggle theme
            if ((e.ctrlKey || e.metaKey) && e.key === '/') {
                e.preventDefault();
                this.toggleTheme();
            }

            // Escape = Close modals
            if (e.key === 'Escape') {
                this.closeModals();
            }
        });
    }

    setupResponsiveActions() {
        // Handle more actions button
        const moreActionsBtn = document.getElementById('more-actions');
        if (moreActionsBtn) {
            moreActionsBtn.addEventListener('click', () => {
                this.toggleMoreActions();
            });
        }

        // Handle window resize to manage responsive layout
        window.addEventListener('resize', () => {
            this.handleResize();
        });

        // Initial responsive setup
        this.handleResize();
    }

    toggleMoreActions() {
        const editorActions = document.getElementById('editor-actions');
        if (editorActions) {
            editorActions.classList.toggle('expanded');

            const moreBtn = document.getElementById('more-actions');
            if (editorActions.classList.contains('expanded')) {
                moreBtn.textContent = '⋯ Less';
            } else {
                moreBtn.textContent = '⋯ More';
            }
        }
    }

    handleResize() {
        const width = window.innerWidth;
        const moreActionsBtn = document.getElementById('more-actions');
        const editorActions = document.getElementById('editor-actions');

        if (width <= 480) {
            // Very small screens - show more button
            if (moreActionsBtn) {
                moreActionsBtn.style.display = 'block';
            }
        } else {
            // Larger screens - hide more button and expand actions
            if (moreActionsBtn) {
                moreActionsBtn.style.display = 'none';
            }
            if (editorActions) {
                editorActions.classList.remove('expanded');
            }
        }
    }

    async setupTheme() {
        try {
            // Load theme from preferences using Wails
            const savedTheme = await window.go.main.App.GetPreference('ui.theme', 'dark');
            this.setTheme(savedTheme);
        } catch (error) {
            console.error('Error loading theme preference:', error);
            this.setTheme('dark'); // fallback
        }

        // Remove theme initialization class after DOM is fully loaded
        setTimeout(() => {
            document.documentElement.classList.remove('theme-initializing');
        }, 100);
    }

    setTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        document.body.dataset.theme = theme;
        document.documentElement.style.setProperty('color-scheme', theme === 'dark' ? 'dark' : 'light');

        // Save to preferences
        window.go.main.App.SetPreference('ui.theme', theme).catch(console.error);

        const themeIcon = document.getElementById('theme-icon');
        if (themeIcon) {
            themeIcon.textContent = theme === 'dark' ? '☀️' : '🌙';
        }
    }

    toggleTheme() {
        const currentTheme = document.body.dataset.theme;
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        this.setTheme(newTheme);
    }

    async loadNotesAndTags() {
        try {
            // Load notes and tags in parallel
            const [notes, tags] = await Promise.all([
                window.go.main.App.ListNotes(50, 0),
                window.go.main.App.GetAllTags()
            ]);

            this.renderNotes(notes);
            this.renderTags(tags);
            this.updateStats(notes.length);

            // Show welcome screen if no notes
            if (notes.length === 0) {
                this.showWelcomeScreen();
            }
        } catch (error) {
            console.error('Error loading notes and tags:', error);
            this.showError('Error', 'Failed to load notes and tags');
        }
    }

    renderNotes(notes) {
        const notesContainer = document.getElementById('notes-list');
        if (!notesContainer) return;

        if (notes.length === 0) {
            notesContainer.innerHTML = `
                <div class="empty-state">
                    <p>No notes yet. Create your first note!</p>
                </div>
            `;
            return;
        }

        // Efficient DOM updates: only update changed notes
        this.updateNotesList(notes, notesContainer);
    }

    updateNotesList(notes, container) {
        // First, check if we need to clear empty state
        const emptyState = container.querySelector('.empty-state');
        if (emptyState && notes.length > 0) {
            emptyState.remove();
        }

        // Create a map of existing notes for efficient lookup
        const existingNotes = new Map();
        const existingElements = container.querySelectorAll('.note-item');

        existingElements.forEach(element => {
            const noteId = parseInt(element.dataset.noteId);
            if (!isNaN(noteId)) {
                existingNotes.set(noteId, element);
            }
        });

        // Create a document fragment for batch DOM updates
        const fragment = document.createDocumentFragment();
        const notesToKeep = new Set();

        // Process each note
        notes.forEach((note, index) => {
            notesToKeep.add(note.id);
            const existingElement = existingNotes.get(note.id);

            if (existingElement) {
                // Update existing element if needed
                this.updateNoteElement(existingElement, note);
                // Move to correct position if needed
                if (existingElement.previousElementSibling !== (index > 0 ? fragment.lastElementChild : null)) {
                    fragment.appendChild(existingElement);
                }
            } else {
                // Create new element
                const noteEl = this.createNoteElement(note);
                fragment.appendChild(noteEl);
            }
        });

        // Remove notes that are no longer in the list
        existingElements.forEach(element => {
            const noteId = parseInt(element.dataset.noteId);
            if (!notesToKeep.has(noteId)) {
                element.remove();
            }
        });

        // Append new/moved elements
        if (fragment.hasChildNodes()) {
            container.appendChild(fragment);
        }
    }

    updateNoteElement(element, note) {
        // Only update if data has changed
        const currentTitle = element.querySelector('.note-title').textContent;
        const currentTags = element.dataset.tags;
        const newTags = note.tags ? note.tags.join(' ') : '';

        if (currentTitle !== note.title) {
            element.querySelector('.note-title').textContent = note.title;
        }

        if (currentTags !== newTags) {
            element.dataset.tags = newTags;
            // Update tags display
            const tagsElement = element.querySelector('.note-tags');
            if (tagsElement) {
                if (note.tags && note.tags.length > 0) {
                    tagsElement.innerHTML = note.tags.map(tag => `<span class="tag">${tag}</span>`).join('');
                } else {
                    tagsElement.innerHTML = '';
                }
            }
        }

        // Update preview if content length suggests it might have changed
        const preview = note.content.length > 100 ? note.content.substring(0, 100) + '...' : note.content;
        const currentPreview = element.querySelector('.note-preview').textContent;
        if (currentPreview !== preview) {
            element.querySelector('.note-preview').textContent = preview;
        }
    }

    renderTags(tags) {
        const tagFilter = document.getElementById('tag-filter');
        if (!tagFilter) return;

        // Clear existing options except "All Tags"
        tagFilter.innerHTML = '<option value="">All Tags</option>';

        tags.forEach(tag => {
            const option = document.createElement('option');
            option.value = tag.name;
            option.textContent = tag.name;
            tagFilter.appendChild(option);
        });
    }

    createNoteElement(note) {
        const noteItem = document.createElement('div');
        noteItem.className = 'note-item';
        noteItem.dataset.noteId = note.id;
        noteItem.dataset.tags = note.tags ? note.tags.join(' ') : '';

        const preview = note.content.length > 100 ? note.content.substring(0, 100) + '...' : note.content;
        const createdDate = new Date(note.created_at).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });

        let tagsHtml = '';
        if (note.tags && note.tags.length > 0) {
            tagsHtml = '<div class="note-tags">' +
                note.tags.map(tag => `<span class="tag">${tag}</span>`).join('') +
                '</div>';
        }

        noteItem.innerHTML = `
            <div class="note-title">${note.title}</div>
            <div class="note-preview">${preview}</div>
            <div class="note-meta">
                <span class="note-date">${createdDate}</span>
                ${tagsHtml}
            </div>
        `;

        // Don't add individual event listeners - use event delegation instead
        // This is handled in setupEventListeners via delegation

        return noteItem;
    }

    showWelcomeScreen() {
        document.getElementById('welcome-screen').style.display = 'flex';
        document.getElementById('note-editor').style.display = 'none';
    }

    showNoteEditor() {
        document.getElementById('welcome-screen').style.display = 'none';
        document.getElementById('note-editor').style.display = 'block';
    }

    async loadNote(noteId) {
        try {
            const note = await window.go.main.App.GetNote(noteId);

            this.currentNoteId = note.id;
            this.isNewNote = false;

            // Populate the form fields
            const titleInput = document.getElementById('note-title');
            const contentTextarea = document.getElementById('note-content');
            const tagsInput = document.getElementById('note-tags');

            if (titleInput) titleInput.value = note.title || '';
            if (contentTextarea) contentTextarea.value = note.content || '';
            if (tagsInput && note.tags) tagsInput.value = note.tags.join(', ');

            // Update tags display
            if (note.tags) {
                this.updateCurrentTags(note.tags);
            }

            // Show delete button for existing notes
            document.getElementById('delete-note').style.display = 'inline-block';

            this.updateDocumentTitle();
            this.showNoteEditor();

            // Update preview if enhanced editor is available
            if (this.enhancedEditor) {
                this.enhancedEditor.updatePreview();
            }

            // Mark all notes as inactive and current as active
            document.querySelectorAll('.note-item').forEach(item => {
                item.classList.remove('active');
                if (parseInt(item.dataset.noteId) === noteId) {
                    item.classList.add('active');
                }
            });

        } catch (error) {
            console.error('Error loading note:', error);
            this.showError('Error', 'Failed to load note');
        }
    }

    createNewNote() {
        this.currentNoteId = null;
        this.isNewNote = true;

        // Clear the form
        document.getElementById('note-title').value = '';
        document.getElementById('note-content').value = '';
        document.getElementById('note-tags').value = '';
        document.getElementById('current-tags').innerHTML = '';

        // Hide delete button for new notes
        document.getElementById('delete-note').style.display = 'none';

        // Clear active note selection
        document.querySelectorAll('.note-item').forEach(item => {
            item.classList.remove('active');
        });

        this.updateDocumentTitle();
        this.showNoteEditor();

        // Focus on title input
        document.getElementById('note-title').focus();

        // Update preview
        if (this.enhancedEditor) {
            this.enhancedEditor.updatePreview();
        }
    }

    async saveCurrentNote() {
        const titleInput = document.getElementById('note-title');
        const tagsInput = document.getElementById('note-tags');

        if (!titleInput.value.trim()) {
            this.showError('Validation Error', 'Please enter a title for your note');
            titleInput.focus();
            return;
        }

        // Get content from enhanced editor
        let content = '';
        if (this.enhancedEditor) {
            content = this.enhancedEditor.getContent();
        } else {
            const contentTextarea = document.getElementById('note-content');
            content = contentTextarea ? contentTextarea.value : '';
        }

        const tags = this.parseTags(tagsInput.value);

        try {
            let note;
            if (this.isNewNote || !this.currentNoteId) {
                // Create new note
                note = await window.go.main.App.CreateNote(titleInput.value, content, tags);
                this.currentNoteId = note.id;
                this.isNewNote = false;
                document.getElementById('delete-note').style.display = 'inline-block';
                this.showSuccess('Note created successfully');
            } else {
                // Update existing note
                note = await window.go.main.App.UpdateNote(this.currentNoteId, titleInput.value, content, tags);
                this.showSuccess('Note saved successfully');
            }

            this.markSaved();
            this.updateNoteTags(note.tags);

            // Efficiently update the notes list instead of full reload
            this.updateNoteInList(note);

        } catch (error) {
            console.error('Error saving note:', error);
            this.showError('Error', 'Failed to save note: ' + error.message);
        }
    }

    async deleteCurrentNote() {
        if (!this.currentNoteId) return;

        const titleInput = document.getElementById('note-title');
        const noteTitle = titleInput.value || 'Untitled Note';

        if (confirm(`Are you sure you want to delete "${noteTitle}"? This action cannot be undone.`)) {
            try {
                const deletedNoteId = this.currentNoteId;
                await window.go.main.App.DeleteNote(deletedNoteId);
                this.showSuccess('Note deleted successfully');

                // Reset to welcome screen
                this.showWelcomeScreen();
                this.currentNoteId = null;
                this.isNewNote = false;

                // Remove note from list efficiently instead of full reload
                this.removeNoteFromList(deletedNoteId);

            } catch (error) {
                console.error('Error deleting note:', error);
                this.showError('Error', 'Failed to delete note: ' + error.message);
            }
        }
    }

    async performSearch(query) {
        if (!query.trim()) {
            this.loadNotesAndTags();
            return;
        }

        try {
            const notes = await window.go.main.App.SearchNotes(query, true, 20);
            this.renderNotes(notes);
        } catch (error) {
            console.error('Error searching notes:', error);
            this.showError('Error', 'Search failed: ' + error.message);
        }
    }

    filterByTag(tag) {
        // Cache the notes container to avoid repeated queries
        const notesContainer = document.getElementById('notes-list');
        if (!notesContainer) return;

        // Use CSS classes for better performance than style.display
        const noteItems = notesContainer.querySelectorAll('.note-item');

        if (!tag) {
            // Show all notes
            noteItems.forEach(item => {
                item.classList.remove('hidden');
            });
        } else {
            const lowerTag = tag.toLowerCase();
            noteItems.forEach(item => {
                const noteTags = item.dataset.tags.toLowerCase();
                if (noteTags.includes(lowerTag)) {
                    item.classList.remove('hidden');
                } else {
                    item.classList.add('hidden');
                }
            });
        }
    }

    async autoTagNote() {
        if (!this.currentNoteId) return;

        const autoTagBtn = document.getElementById('auto-tag-btn');
        if (autoTagBtn) {
            autoTagBtn.disabled = true;
            autoTagBtn.textContent = '🤖 Generating...';
        }

        try {
            const suggestedTags = await window.go.main.App.SuggestTags(this.currentNoteId);

            if (suggestedTags && suggestedTags.length > 0) {
                this.addSuggestedTags(suggestedTags);
                this.showSuccess(`Added ${suggestedTags.length} auto-generated tags`);
            } else {
                this.showInfo('No tags suggested');
            }
        } catch (error) {
            console.error('Error auto-tagging:', error);
            this.showError('Error', 'Auto-tagging failed: ' + error.message);
        } finally {
            if (autoTagBtn) {
                autoTagBtn.disabled = false;
                autoTagBtn.textContent = '🏷️ Auto-tag';
            }
        }
    }

    async analyzeNote() {
        if (!this.currentNoteId) return;

        const analyzeBtn = document.getElementById('analyze-btn');
        if (analyzeBtn) {
            analyzeBtn.disabled = true;
            analyzeBtn.textContent = '🤖 Analyzing...';
        }

        try {
            const result = await window.go.main.App.AnalyzeNote(this.currentNoteId, '');

            this.showModal('Analysis Result', result.analysis);

        } catch (error) {
            console.error('Error analyzing note:', error);
            this.showError('Error', 'Analysis failed: ' + error.message);
        } finally {
            if (analyzeBtn) {
                analyzeBtn.disabled = false;
                analyzeBtn.textContent = '🤖 Analyze';
            }
        }
    }

    // Helper methods

    parseTags(tagsStr) {
        if (!tagsStr) return [];

        return tagsStr.split(',')
            .map(tag => tag.trim())
            .filter(tag => tag.length > 0);
    }

    addSuggestedTags(suggestedTags) {
        const tagsInput = document.getElementById('note-tags');
        if (!tagsInput) return;

        const currentTags = this.parseTags(tagsInput.value);
        const newTags = [...new Set([...currentTags, ...suggestedTags])];

        tagsInput.value = newTags.join(', ');
        this.updateCurrentTags(newTags);
        this.markUnsaved();
    }

    updateCurrentTags(tags) {
        const currentTagsContainer = document.getElementById('current-tags');
        if (!currentTagsContainer) return;

        currentTagsContainer.innerHTML = '';

        tags.forEach(tag => {
            if (tag.trim()) {
                const tagElement = document.createElement('span');
                tagElement.className = 'tag removable';
                tagElement.dataset.tag = tag.trim();
                tagElement.innerHTML = `${tag.trim()} <span class="tag-remove">×</span>`;
                currentTagsContainer.appendChild(tagElement);
            }
        });
    }

    updateNoteTags(tags) {
        const tagsInput = document.getElementById('note-tags');
        if (tagsInput && tags) {
            tagsInput.value = tags.join(', ');
            this.updateCurrentTags(tags);
        }
    }

    removeTag(tagToRemove) {
        const tagsInput = document.getElementById('note-tags');
        if (!tagsInput) return;

        const currentTags = this.parseTags(tagsInput.value).filter(tag => tag !== tagToRemove);
        tagsInput.value = currentTags.join(', ');
        this.updateCurrentTags(currentTags);
        this.markUnsaved();
    }

    updateDocumentTitle() {
        const titleInput = document.getElementById('note-title');
        if (titleInput && titleInput.value) {
            document.title = `${titleInput.value} - ML Notes`;
        } else {
            document.title = 'ML Notes';
        }
    }

    updateStats(noteCount) {
        const statsEl = document.getElementById('stats-notes');
        if (statsEl) {
            statsEl.textContent = `${noteCount} notes`;
        }
    }

    updateNoteInList(note) {
        const notesContainer = document.getElementById('notes-list');
        if (!notesContainer) return;

        const existingElement = notesContainer.querySelector(`[data-note-id="${note.id}"]`);
        if (existingElement) {
            // Update existing note
            this.updateNoteElement(existingElement, note);
        } else {
            // Add new note at the top
            const newElement = this.createNoteElement(note);
            const firstNote = notesContainer.querySelector('.note-item');
            if (firstNote) {
                notesContainer.insertBefore(newElement, firstNote);
            } else {
                // Replace empty state with first note
                notesContainer.innerHTML = '';
                notesContainer.appendChild(newElement);
            }
        }
    }

    removeNoteFromList(noteId) {
        const notesContainer = document.getElementById('notes-list');
        if (!notesContainer) return;

        const noteElement = notesContainer.querySelector(`[data-note-id="${noteId}"]`);
        if (noteElement) {
            noteElement.remove();

            // Check if we need to show empty state
            const remainingNotes = notesContainer.querySelectorAll('.note-item');
            if (remainingNotes.length === 0) {
                notesContainer.innerHTML = `
                    <div class="empty-state">
                        <p>No notes yet. Create your first note!</p>
                    </div>
                `;
            }
        }
    }

    markUnsaved() {
        this.unsavedChanges = true;
        const saveBtn = document.getElementById('save-note');
        if (saveBtn) {
            saveBtn.textContent = 'Save*';
            saveBtn.classList.add('unsaved');
        }
    }

    markSaved() {
        this.unsavedChanges = false;
        const saveBtn = document.getElementById('save-note');
        if (saveBtn) {
            saveBtn.textContent = 'Save';
            saveBtn.classList.remove('unsaved');
        }
    }

    // UI Helper methods

    showModal(title, message) {
        const modal = document.getElementById('modal-overlay');
        const modalTitle = document.getElementById('modal-title');
        const modalMessage = document.getElementById('modal-message');

        modalTitle.textContent = title;
        modalMessage.textContent = message;
        modal.style.display = 'flex';
    }

    showError(title, message) {
        this.showModal(title, message);
    }

    showSuccess(message) {
        this.showNotification(message, 'success');
    }

    showInfo(message) {
        this.showNotification(message, 'info');
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;

        document.body.appendChild(notification);

        // Animate in
        setTimeout(() => {
            notification.classList.add('show');
        }, 100);

        // Remove after delay
        setTimeout(() => {
            notification.classList.remove('show');
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 3000);
    }

    closeModals() {
        const overlay = document.getElementById('modal-overlay');
        if (overlay) {
            overlay.style.display = 'none';
        }
    }

    setupMarkdownPreview() {
        // Setup handled by EnhancedEditor
    }

    initializeSlidingPanels() {
        const sideBySideEditor = document.getElementById('side-by-side-editor');
        const hideEditorBtn = document.getElementById('hide-editor');
        const hidePreviewBtn = document.getElementById('hide-preview');

        if (!sideBySideEditor) return;

        // Hide editor panel
        hideEditorBtn?.addEventListener('click', () => {
            sideBySideEditor.classList.add('hide-editor');
            sideBySideEditor.classList.remove('hide-preview');
        });

        // Hide preview panel
        hidePreviewBtn?.addEventListener('click', () => {
            sideBySideEditor.classList.add('hide-preview');
            sideBySideEditor.classList.remove('hide-editor');
        });
    }

    // Settings and Initialization Methods

    async checkInitialization() {
        try {
            const isInitialized = await window.go.main.App.IsConfigInitialized();
            return isInitialized;
        } catch (error) {
            console.error('Error checking initialization:', error);
            return false;
        }
    }

    showInitializationPage() {
        const mainContent = document.querySelector('.main-content');
        if (mainContent) mainContent.style.display = 'none';
        document.getElementById('init-page').style.display = 'block';
        document.getElementById('settings-page').style.display = 'none';
    }

    hideInitializationPage() {
        document.getElementById('init-page').style.display = 'none';
        const mainContent = document.querySelector('.main-content');
        if (mainContent) mainContent.style.display = 'block';
    }

    async initializeApp() {
        const dataDir = document.getElementById('init-data-dir').value.trim();
        const ollamaEndpoint = document.getElementById('init-ollama-endpoint').value.trim();

        if (!dataDir) {
            this.showError('Validation Error', 'Data directory is required');
            return;
        }

        if (!ollamaEndpoint) {
            this.showError('Validation Error', 'Ollama endpoint is required');
            return;
        }

        try {
            await window.go.main.App.InitializeConfig(dataDir, ollamaEndpoint);
            this.showSuccess('Application initialized successfully! Please restart the application.');

            // Hide init page and show main content after a short delay
            setTimeout(() => {
                this.hideInitializationPage();
                this.init(); // Reinitialize the app
            }, 2000);
        } catch (error) {
            console.error('Error initializing app:', error);
            this.showError('Initialization Error', 'Failed to initialize: ' + error.message);
        }
    }

    showSettingsPage() {
        console.log('showSettingsPage called');

        // Remove any existing settings overlay
        const existing = document.getElementById('dynamic-settings-overlay');
        if (existing) {
            document.body.removeChild(existing);
        }

        // Create dynamic settings overlay
        const settingsOverlay = document.createElement('div');
        settingsOverlay.id = 'dynamic-settings-overlay';

        // Create the settings HTML content
        settingsOverlay.innerHTML = `
            <div style="background: var(--paper-bg); color: var(--text-primary); max-width: 800px; margin: 0 auto; padding: 20px; height: 100%; overflow-y: auto; box-sizing: border-box;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 30px;">
                    <h2 style="margin: 0; font-size: 2rem;">Settings</h2>
                    <button id="dynamic-close-settings" style="background: #666; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer;">Close</button>
                </div>

                <div style="margin-bottom: 30px;">
                    <h3>General Settings</h3>
                    <div style="margin-bottom: 15px;">
                        <label>Data Directory:</label><br>
                        <input type="text" id="dynamic-data-directory" style="width: 100%; padding: 8px; margin-top: 5px; border: 1px solid #666; border-radius: 4px;">
                    </div>
                    <div style="margin-bottom: 15px;">
                        <label><input type="checkbox" id="dynamic-debug-mode" style="margin-right: 8px;"> Debug Mode</label>
                    </div>
                    <div style="margin-bottom: 15px;">
                        <label>UI Theme:</label><br>
                        <select id="dynamic-webui-theme" style="width: 100%; padding: 8px; margin-top: 5px; border: 1px solid #666; border-radius: 4px;">
                            <option value="dark">Dark</option>
                            <option value="light">Light</option>
                            <option value="auto">Auto</option>
                        </select>
                    </div>
                </div>

                <div style="margin-bottom: 30px;">
                    <h3>AI Services</h3>
                    <div style="margin-bottom: 15px;">
                        <label>Ollama Endpoint:</label><br>
                        <input type="url" id="dynamic-ollama-endpoint" placeholder="http://localhost:11434" style="width: 100%; padding: 8px; margin-top: 5px; border: 1px solid #666; border-radius: 4px;">
                    </div>
                    <div style="margin-bottom: 15px;">
                        <label>Lil-Rag URL:</label><br>
                        <input type="url" id="dynamic-lilrag-url" placeholder="http://localhost:12121" style="width: 100%; padding: 8px; margin-top: 5px; border: 1px solid #666; border-radius: 4px;">
                    </div>
                    <div style="margin-bottom: 15px;">
                        <label><input type="checkbox" id="dynamic-enable-summarization" style="margin-right: 8px;"> Enable Summarization</label>
                    </div>
                    <div style="margin-bottom: 15px;">
                        <label>Summarization Model:</label><br>
                        <input type="text" id="dynamic-summarization-model" placeholder="llama3.2:latest" style="width: 100%; padding: 8px; margin-top: 5px; border: 1px solid #666; border-radius: 4px;">
                    </div>
                    <div style="margin-bottom: 15px;">
                        <label><input type="checkbox" id="dynamic-enable-auto-tagging" style="margin-right: 8px;"> Enable Auto-Tagging</label>
                    </div>
                    <div style="margin-bottom: 15px;">
                        <label>Auto-Tag Model:</label><br>
                        <input type="text" id="dynamic-auto-tag-model" placeholder="llama3.2:latest" style="width: 100%; padding: 8px; margin-top: 5px; border: 1px solid #666; border-radius: 4px;">
                    </div>
                    <div style="margin-bottom: 15px;">
                        <label>Max Auto Tags:</label><br>
                        <input type="number" id="dynamic-max-auto-tags" min="1" max="20" value="5" style="width: 100%; padding: 8px; margin-top: 5px; border: 1px solid #666; border-radius: 4px;">
                    </div>
                </div>

                <div style="margin-bottom: 30px;">
                    <h3>Advanced Settings</h3>
                    <div style="margin-bottom: 15px;">
                        <label>External Editor:</label><br>
                        <input type="text" id="dynamic-editor" placeholder="code --wait" style="width: 100%; padding: 8px; margin-top: 5px; border: 1px solid #666; border-radius: 4px;">
                        <small style="color: #888; font-size: 12px;">Command to launch external editor (e.g., code --wait, vim, emacs)</small>
                    </div>
                    <div style="margin-bottom: 15px;">
                        <label>Custom CSS:</label><br>
                        <textarea id="dynamic-webui-custom-css" placeholder="/* Custom CSS rules */" rows="4" style="width: 100%; padding: 8px; margin-top: 5px; border: 1px solid #666; border-radius: 4px; font-family: monospace; resize: vertical;"></textarea>
                    </div>
                </div>

                <div style="margin-bottom: 30px;">
                    <h3>GitHub Integration</h3>
                    <div style="margin-bottom: 15px;">
                        <label>GitHub Owner:</label><br>
                        <input type="text" id="dynamic-github-owner" placeholder="username" style="width: 100%; padding: 8px; margin-top: 5px; border: 1px solid #666; border-radius: 4px;">
                    </div>
                    <div style="margin-bottom: 15px;">
                        <label>GitHub Repository:</label><br>
                        <input type="text" id="dynamic-github-repo" placeholder="repository-name" style="width: 100%; padding: 8px; margin-top: 5px; border: 1px solid #666; border-radius: 4px;">
                    </div>
                </div>

                <div style="text-align: center;">
                    <button id="dynamic-save-settings" style="background: #007bff; color: white; border: none; padding: 12px 30px; border-radius: 4px; cursor: pointer; margin-right: 10px;">Save Settings</button>
                    <button id="dynamic-test-ollama" style="background: #6c757d; color: white; border: none; padding: 12px 30px; border-radius: 4px; cursor: pointer;">Test Ollama</button>
                </div>
            </div>
        `;

        // Style the overlay
        settingsOverlay.style.position = 'fixed';
        settingsOverlay.style.top = '0';
        settingsOverlay.style.left = '0';
        settingsOverlay.style.width = '100vw';
        settingsOverlay.style.height = '100vh';
        settingsOverlay.style.zIndex = '99999';
        settingsOverlay.style.backgroundColor = 'var(--paper-bg)';
        settingsOverlay.style.display = 'block';

        // Add to body
        document.body.appendChild(settingsOverlay);

        // Load settings data
        this.loadDynamicSettings();

        // Add event listeners
        this.setupDynamicSettingsListeners(settingsOverlay);
    }

    hideSettingsPage() {
        document.getElementById('settings-page').style.display = 'none';
        const mainContent = document.querySelector('.main-content');
        if (mainContent) mainContent.style.display = 'block';
    }

    async loadSettings() {
        console.log('loadSettings method called');
        try {
            console.log('Checking if window.go exists:', !!window.go);
            console.log('Checking if window.go.main exists:', !!(window.go && window.go.main));
            console.log('Checking if GetConfig exists:', !!(window.go && window.go.main && window.go.main.App && window.go.main.App.GetConfig));

            console.log('About to call GetConfig...');
            window.runtime.LogInfo('JavaScript: About to call GetConfig...');
            const config = await window.go.main.App.GetConfig();
            console.log('Config loaded successfully:', config);
            window.runtime.LogInfo('JavaScript: Config loaded successfully');

            // Check if elements exist before setting values
            const elements = {
                'data-directory': config.data_directory || '',
                'ollama-endpoint': config.ollama_endpoint || '',
                'debug-mode': config.debug || false,
                'summarization-model': config.summarization_model || '',
                'enable-summarization': config.enable_summarization || false,
                'editor': config.editor || '',
                'enable-auto-tagging': config.enable_auto_tagging || false,
                'auto-tag-model': config.auto_tag_model || '',
                'max-auto-tags': config.max_auto_tags || 5,
                'webui-theme': config.webui_theme || 'dark',
                'webui-custom-css': config.webui_custom_css || '',
                'lilrag-url': config.lilrag_url || ''
            };

            // Populate form fields with existence checks
            Object.entries(elements).forEach(([id, value]) => {
                const element = document.getElementById(id);
                if (element) {
                    if (element.type === 'checkbox') {
                        element.checked = Boolean(value);
                    } else {
                        element.value = value;
                    }
                } else {
                    console.warn(`Element with id '${id}' not found`);
                }
            });

        } catch (error) {
            console.error('Error loading settings:', error);
            this.showError('Error', 'Failed to load settings: ' + error.message);
        }
    }

    async saveSettings() {
        try {
            const updates = {
                data_directory: document.getElementById('data-directory').value.trim(),
                ollama_endpoint: document.getElementById('ollama-endpoint').value.trim(),
                debug: document.getElementById('debug-mode').checked,
                summarization_model: document.getElementById('summarization-model').value.trim(),
                enable_summarization: document.getElementById('enable-summarization').checked,
                editor: document.getElementById('editor').value.trim(),
                enable_auto_tagging: document.getElementById('enable-auto-tagging').checked,
                auto_tag_model: document.getElementById('auto-tag-model').value.trim(),
                max_auto_tags: parseInt(document.getElementById('max-auto-tags').value) || 5,
                webui_theme: document.getElementById('webui-theme').value,
                webui_custom_css: document.getElementById('webui-custom-css').value.trim(),
                lilrag_url: document.getElementById('lilrag-url').value.trim()
            };

            await window.go.main.App.UpdateConfig(updates);
            this.showSuccess('Settings saved successfully! Some changes may require an application restart.');

        } catch (error) {
            console.error('Error saving settings:', error);
            this.showError('Error', 'Failed to save settings: ' + error.message);
        }
    }

    async testOllamaConnection() {
        const testBtn = document.getElementById('test-ollama');
        const originalText = testBtn.textContent;

        try {
            testBtn.textContent = 'Testing...';
            testBtn.disabled = true;

            const result = await window.go.main.App.TestOllamaConnection();

            if (result.success) {
                this.showSuccess('Ollama Connection Successful: ' + result.message);
            } else {
                this.showError('Ollama Connection Failed', result.error);
            }

        } catch (error) {
            console.error('Error testing Ollama connection:', error);
            this.showError('Error', 'Failed to test Ollama connection: ' + error.message);
        } finally {
            testBtn.textContent = originalText;
            testBtn.disabled = false;
        }
    }

    // Dynamic Settings Methods
    async loadDynamicSettings() {
        try {
            const config = await window.go.main.App.GetConfig();

            // Populate dynamic form fields
            const elements = {
                'dynamic-data-directory': config.data_directory || '',
                'dynamic-ollama-endpoint': config.ollama_endpoint || '',
                'dynamic-debug-mode': config.debug || false,
                'dynamic-webui-theme': config.webui_theme || 'dark',
                'dynamic-lilrag-url': config.lilrag_url || '',
                'dynamic-enable-summarization': config.enable_summarization || false,
                'dynamic-summarization-model': config.summarization_model || '',
                'dynamic-enable-auto-tagging': config.enable_auto_tagging || false,
                'dynamic-auto-tag-model': config.auto_tag_model || '',
                'dynamic-max-auto-tags': config.max_auto_tags || 5,
                'dynamic-editor': config.editor || '',
                'dynamic-webui-custom-css': config.webui_custom_css || '',
                'dynamic-github-owner': config.github_owner || '',
                'dynamic-github-repo': config.github_repo || ''
            };

            Object.entries(elements).forEach(([id, value]) => {
                const element = document.getElementById(id);
                if (element) {
                    if (element.type === 'checkbox') {
                        element.checked = Boolean(value);
                    } else {
                        element.value = value;
                    }
                }
            });

        } catch (error) {
            console.error('Error loading dynamic settings:', error);
            alert('Failed to load settings: ' + error.message);
        }
    }

    setupDynamicSettingsListeners(overlay) {
        // Close button
        const closeBtn = document.getElementById('dynamic-close-settings');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => {
                document.body.removeChild(overlay);
            });
        }

        // Save settings button
        const saveBtn = document.getElementById('dynamic-save-settings');
        if (saveBtn) {
            saveBtn.addEventListener('click', () => {
                this.saveDynamicSettings();
            });
        }

        // Test Ollama button
        const testBtn = document.getElementById('dynamic-test-ollama');
        if (testBtn) {
            testBtn.addEventListener('click', () => {
                this.testDynamicOllamaConnection();
            });
        }
    }

    async saveDynamicSettings() {
        try {
            const updates = {
                data_directory: document.getElementById('dynamic-data-directory').value.trim(),
                ollama_endpoint: document.getElementById('dynamic-ollama-endpoint').value.trim(),
                debug: document.getElementById('dynamic-debug-mode').checked,
                webui_theme: document.getElementById('dynamic-webui-theme').value,
                lilrag_url: document.getElementById('dynamic-lilrag-url').value.trim(),
                enable_summarization: document.getElementById('dynamic-enable-summarization').checked,
                summarization_model: document.getElementById('dynamic-summarization-model').value.trim(),
                enable_auto_tagging: document.getElementById('dynamic-enable-auto-tagging').checked,
                auto_tag_model: document.getElementById('dynamic-auto-tag-model').value.trim(),
                max_auto_tags: parseInt(document.getElementById('dynamic-max-auto-tags').value) || 5,
                editor: document.getElementById('dynamic-editor').value.trim(),
                webui_custom_css: document.getElementById('dynamic-webui-custom-css').value.trim(),
                github_owner: document.getElementById('dynamic-github-owner').value.trim(),
                github_repo: document.getElementById('dynamic-github-repo').value.trim()
            };

            await window.go.main.App.UpdateConfig(updates);
            alert('Settings saved successfully! Some changes may require an application restart.');

        } catch (error) {
            console.error('Error saving dynamic settings:', error);
            alert('Failed to save settings: ' + error.message);
        }
    }

    async testDynamicOllamaConnection() {
        const testBtn = document.getElementById('dynamic-test-ollama');
        const originalText = testBtn.textContent;

        try {
            testBtn.textContent = 'Testing...';
            testBtn.disabled = true;

            const result = await window.go.main.App.TestOllamaConnection();

            if (result.success) {
                alert('Ollama Connection Successful: ' + result.message);
            } else {
                alert('Ollama Connection Failed: ' + result.error);
            }

        } catch (error) {
            console.error('Error testing Ollama connection:', error);
            alert('Failed to test Ollama connection: ' + error.message);
        } finally {
            testBtn.textContent = originalText;
            testBtn.disabled = false;
        }
    }
}

// Enhanced Editor - Simple layered preview (same as original)
class EnhancedEditor {
    constructor(textarea, preview) {
        this.textarea = textarea;
        this.preview = preview;
        this.renderTimeout = null;

        this.init();
    }

    init() {
        this.textarea.addEventListener('input', () => this.handleInput());
        this.textarea.addEventListener('scroll', () => this.syncScroll());
        this.updatePreview();
    }

    handleInput() {
        clearTimeout(this.renderTimeout);
        // Increased debounce delay for better performance
        this.renderTimeout = setTimeout(() => {
            this.updatePreview();
        }, 300);
    }

    syncScroll() {
        const scrollPercentage = this.textarea.scrollTop / (this.textarea.scrollHeight - this.textarea.clientHeight);
        const targetScrollTop = scrollPercentage * (this.preview.scrollHeight - this.preview.clientHeight);
        this.preview.scrollTop = targetScrollTop;
    }

    updatePreview() {
        const markdown = this.textarea.value;
        if (!markdown.trim()) {
            this.preview.innerHTML = '';
            return;
        }

        try {
            let html = marked.parse(markdown);

            if (window.DOMPurify) {
                html = DOMPurify.sanitize(html);
            }

            this.preview.innerHTML = html;
            this.syncScroll();
        } catch (error) {
            console.error('Error rendering markdown:', error);
        }
    }

    getContent() {
        return this.textarea.value;
    }
}

// Initialize the app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM loaded');
    window.runtime.LogInfo('JavaScript: DOM loaded');

    // Wait for Wails runtime to be available
    window.addEventListener('wails:ready', () => {
        console.log('Wails ready event fired');
        window.runtime.LogInfo('JavaScript: Wails ready event fired');
        new MLNotesWailsApp();
    });

    // Fallback in case wails:ready doesn't fire
    setTimeout(() => {
        console.log('Fallback timeout reached');
        window.runtime.LogInfo('JavaScript: Fallback timeout reached');
        if (window.go && window.go.main && window.go.main.App) {
            console.log('Creating MLNotesWailsApp via fallback');
            window.runtime.LogInfo('JavaScript: Creating MLNotesWailsApp via fallback');
            new MLNotesWailsApp();
        } else {
            console.log('window.go not available in fallback');
            window.runtime.LogError('JavaScript: window.go not available in fallback');
        }
    }, 1000);
});