(function() {
    'use strict';
    
    let cachedElements = {};
    
    function getCachedElement(selector, parent = document) {
        if (!cachedElements[selector]) {
            cachedElements[selector] = parent.querySelector(selector);
        }
        return cachedElements[selector];
    }
    
    // Check current language
    const isKorean = window.location.pathname.startsWith('/ko/');
    
    // Language switching function
    function switchLanguage(lang) {
        const currentPath = window.location.pathname;
        let newPath;
        
        if (lang === 'ko') {
            if (currentPath.startsWith('/ko/')) return;
            newPath = '/ko' + currentPath;
        } else {
            if (currentPath.startsWith('/ko/')) {
                newPath = currentPath.substring(3);
            } else return;
        }
        
        // Set cookie and navigate to new page
        document.cookie = `preferred_language=${lang}; path=/; max-age=31536000`;
        window.location.href = newPath;
    }
    
    // Update language selector
    function updateLanguageSelector() {
        const langButton = getCachedElement('.md-header__button[aria-label="Select language"]');
        if (!langButton) return;
        
        const currentLangText = isKorean ? '한국어 (Korean)' : 'English';
        
        // Use CSS classes for style control
        document.body.className = document.body.className.replace(/(^|\s)(is-korean|is-english)(\s|$)/g, ' ').trim() + 
                                  (isKorean ? ' is-korean' : ' is-english');
        
        // Update existing text or create new one
        let langText = langButton.querySelector('.lang-text');
        if (!langText) {
            const icon = langButton.querySelector('svg');
            if (icon) {
                icon.style.display = 'none';
                langText = document.createElement('span');
                langText.className = 'lang-text';
                langButton.appendChild(langText);
            }
        }
        
        if (langText) {
            langText.textContent = currentLangText;
        }
        
        langButton.classList.add('current-language');
        updateDropdownMenu();
    }
    
    // Update dropdown menu
    function updateDropdownMenu() {
        const dropdown = getCachedElement('.md-select__inner');
        if (!dropdown) return;
        
        const links = dropdown.querySelectorAll('.md-select__link');
        
        links.forEach(link => {
            const href = link.getAttribute('href');
            
            if (href === '/' || href === './') {
                link.textContent = 'English';
                link.classList.toggle('current', !isKorean);
            } else if (href === '/ko/' || href.startsWith('/ko/')) {
                link.textContent = '한국어 (Korean)';
                link.classList.toggle('current', isKorean);
            }
            
            // Add event listener
            if (!link.hasAttribute('data-lang-listener')) {
                link.setAttribute('data-lang-listener', 'true');
                link.addEventListener('click', function(e) {
                    const targetIsKorean = href.startsWith('/ko/');
                    
                    if (isKorean === targetIsKorean) {
                        e.preventDefault();
                        return;
                    }
                    
                    const targetLang = targetIsKorean ? 'ko' : 'en';
                    document.cookie = `preferred_language=${targetLang}; path=/; max-age=31536000`;
                });
            }
        });
    }
    
    // Initialization function
    function init() {
        updateLanguageSelector();
        
        // Detect dynamic content with MutationObserver
        const observer = new MutationObserver(function(mutations) {
            mutations.forEach(function(mutation) {
                if (mutation.type === 'childList' && mutation.addedNodes.length > 0) {
                    // Update only when new language links are added
                    for (let node of mutation.addedNodes) {
                        if (node.nodeType === 1 && node.matches && node.matches('.md-select__link')) {
                            updateDropdownMenu();
                            break;
                        }
                    }
                }
            });
        });
        
        // Observe language selector container
        const selectContainer = getCachedElement('.md-select');
        if (selectContainer) {
            observer.observe(selectContainer, {
                childList: true,
                subtree: true
            });
        }
    }
    
    // Execute immediately when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        // Execute immediately if already loaded
        init();
    }
    
    // Expose global function
    window.switchLanguage = switchLanguage;
    
})(); 