/**
 * Un-tie.me code - Main Application JavaScript
 * 
 * This file implements a modular frontend with robust state management
 * following SOLID principles and composition over inheritance.
 */

// Immediately Invoked Function Expression for encapsulation
(function() {
  'use strict';
  
  // Application namespace
  const UnTieApp = {
    // State management using a simple pub/sub pattern
    state: {
      data: {},
      listeners: {},
      
      // Set state data with immutability
      set: function(key, value) {
        this.data = {
          ...this.data,
          [key]: value
        };
        this.notify(key);
        
        // For debugging
        console.log(`State updated: ${key}`, value);
      },
      
      // Get state data
      get: function(key) {
        return this.data[key];
      },
      
      // Subscribe to state changes
      subscribe: function(key, callback) {
        if (!this.listeners[key]) {
          this.listeners[key] = [];
        }
        this.listeners[key].push(callback);
        
        // Return unsubscribe function for cleanup
        return () => {
          this.listeners[key] = this.listeners[key].filter(cb => cb !== callback);
        };
      },
      
      // Notify subscribers of state changes
      notify: function(key) {
        if (this.listeners[key]) {
          this.listeners[key].forEach(callback => {
            callback(this.data[key]);
          });
        }
      }
    },
    
    // UI Components using composition
    components: {
      toast: {
        show: function(message, type = 'info', duration = 3000) {
          const toast = document.createElement('div');
          toast.className = `toast toast-${type} fade-in`;
          toast.textContent = message;
          
          document.body.appendChild(toast);
          
          setTimeout(() => {
            toast.style.opacity = '0';
            toast.style.transition = 'opacity 0.3s ease';
            
            setTimeout(() => {
              document.body.removeChild(toast);
            }, 300);
          }, duration);
        }
      },
      
      modal: {
        show: function(title, content, actions) {
          // Create modal structure
          const modalOverlay = document.createElement('div');
          modalOverlay.className = 'fixed inset-0 bg-black bg-opacity-50 z-40 flex items-center justify-center';
          
          const modalContent = document.createElement('div');
          modalContent.className = 'bg-white rounded-lg shadow-xl max-w-md w-full mx-4';
          
          // Modal header
          const header = document.createElement('div');
          header.className = 'px-6 py-4 border-b';
          header.innerHTML = `<h3 class="text-lg font-medium">${title}</h3>`;
          
          // Modal body
          const body = document.createElement('div');
          body.className = 'px-6 py-4';
          body.innerHTML = content;
          
          // Modal footer with actions
          const footer = document.createElement('div');
          footer.className = 'px-6 py-4 border-t flex justify-end space-x-2';
          
          // Add actions
          actions.forEach(action => {
            const button = document.createElement('button');
            button.className = action.primary ? 
              'px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700' : 
              'px-4 py-2 border rounded hover:bg-gray-50';
            button.textContent = action.text;
            button.addEventListener('click', () => {
              if (action.callback) action.callback();
              this.hide(modalOverlay);
            });
            footer.appendChild(button);
          });
          
          // Close on overlay click
          modalOverlay.addEventListener('click', (e) => {
            if (e.target === modalOverlay) this.hide(modalOverlay);
          });
          
          // Assemble modal
          modalContent.appendChild(header);
          modalContent.appendChild(body);
          modalContent.appendChild(footer);
          modalOverlay.appendChild(modalContent);
          
          document.body.appendChild(modalOverlay);
        },
        
        hide: function(modalElement) {
          modalElement.classList.add('fade-out');
          setTimeout(() => {
            document.body.removeChild(modalElement);
          }, 300);
        }
      }
    },
    
    // Services for API interactions
    services: {
      api: {
        get: async function(endpoint) {
          try {
            const response = await fetch(endpoint);
            if (!response.ok) throw new Error(`API error: ${response.status}`);
            return await response.json();
          } catch (error) {
            console.error('API GET error:', error);
            UnTieApp.components.toast.show(`Error: ${error.message}`, 'error');
            throw error;
          }
        },
        
        post: async function(endpoint, data) {
          try {
            const response = await fetch(endpoint, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json'
              },
              body: JSON.stringify(data)
            });
            
            if (!response.ok) throw new Error(`API error: ${response.status}`);
            return await response.json();
          } catch (error) {
            console.error('API POST error:', error);
            UnTieApp.components.toast.show(`Error: ${error.message}`, 'error');
            throw error;
          }
        },
        
        put: async function(endpoint, data) {
          try {
            const response = await fetch(endpoint, {
              method: 'PUT',
              headers: {
                'Content-Type': 'application/json'
              },
              body: JSON.stringify(data)
            });
            
            if (!response.ok) throw new Error(`API error: ${response.status}`);
            return await response.json();
          } catch (error) {
            console.error('API PUT error:', error);
            UnTieApp.components.toast.show(`Error: ${error.message}`, 'error');
            throw error;
          }
        },
        
        delete: async function(endpoint) {
          try {
            const response = await fetch(endpoint, {
              method: 'DELETE'
            });
            
            if (!response.ok) throw new Error(`API error: ${response.status}`);
            return await response.json();
          } catch (error) {
            console.error('API DELETE error:', error);
            UnTieApp.components.toast.show(`Error: ${error.message}`, 'error');
            throw error;
          }
        }
      }
    },
    
    // Initialize the application
    init: function() {
      // Setup HTMX events
      document.body.addEventListener('htmx:afterSwap', function(event) {
        // Reinitialize components after HTMX content swap
        UnTieApp.initializeComponents();
      });
      
      // Initialize components on first load
      this.initializeComponents();
      
      console.log('Un-tie.me code application initialized');
    },
    
    // Initialize interactive components
    initializeComponents: function() {
      // Initialize any components that need setup
      // This will be called on initial load and after HTMX swaps
      
      // Example: Setup drag-and-drop for story cards if they exist
      const storyCards = document.querySelectorAll('.story-card');
      if (storyCards.length > 0) {
        // Initialize drag and drop (placeholder for actual implementation)
        console.log('Story cards initialized');
      }
      
      // Example: Initialize architecture canvas if it exists
      const archCanvas = document.getElementById('architecture-canvas');
      if (archCanvas) {
        // Initialize canvas (placeholder for actual implementation)
        console.log('Architecture canvas initialized');
      }
    }
  };
  
  // Initialize the application when DOM is ready
  document.addEventListener('DOMContentLoaded', function() {
    UnTieApp.init();
    
    // Expose to window for debugging only in development
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
      window.UnTieApp = UnTieApp;
    }
  });
})();
