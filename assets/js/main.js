/**
 * Laura & Gerard Wedding - Main JavaScript
 * Lightweight, vanilla JS for countdown, FAQ, and animations
 */

(function() {
    'use strict';

    // ===================
    // Language Detection (runs immediately, before DOMContentLoaded)
    // ===================
    function initLanguageDetection() {
        // Skip if user has manually selected a language before
        if (localStorage.getItem('lang-selected')) return;

        const path = window.location.pathname;

        // Determine current language from URL
        const currentLang = path.includes('/es/') ? 'es' :
                           path.includes('/ca/') ? 'ca' : 'en';

        // Detect browser's preferred language
        // Priority: Catalan > then first supported language found
        const browserLangs = navigator.languages || [navigator.language || 'en'];
        const supportedLangs = browserLangs.map(l => l.toLowerCase().split('-')[0]);
        
        let preferredLang = 'en'; // default fallback
        
        // Always prefer Catalan if it's anywhere in the supported list
        if (supportedLangs.includes('ca')) {
            preferredLang = 'ca';
        } else {
            // Otherwise, use the first supported language
            for (const code of supportedLangs) {
                if (code === 'es') { preferredLang = 'es'; break; }
                if (code === 'en') { preferredLang = 'en'; break; }
            }
        }

        // Redirect if browser language differs from current page language
        if (preferredLang !== currentLang) {
            // Build the new URL preserving the current page path
            // Remove leading slash and language prefix, then add preferred language
            let pagePath = path.replace(/^\//, ''); // Remove leading slash
            pagePath = pagePath.replace(/^(en|es|ca)(\/|$)/, ''); // Remove language prefix if present
            
            const langPrefix = preferredLang === 'en' ? '/' : '/' + preferredLang + '/';
            
            // Preserve query parameters (e.g., ?code=ABC)
            const queryString = window.location.search;
            const newUrl = langPrefix + pagePath + queryString;
            
            window.location.replace(newUrl);
        }
    }

    // Run language detection immediately
    initLanguageDetection();

    // ===================
    // Countdown Timer
    // ===================
    // Wedding: Dec 19, 2026 at 4:00 PM in Copán Ruinas, Honduras (UTC-6)
    const weddingDate = new Date("2026-12-19T16:00:00-06:00").getTime();

    function updateCountdown() {
        const now = new Date().getTime();
        const distance = weddingDate - now;

        if (distance < 0) {
            const timer = document.getElementById("timer");
            if (timer) {
                timer.innerHTML = '<div class="col-span-4 text-2xl font-serif">The celebration has begun!</div>';
            }
            return;
        }

        const days = Math.floor(distance / (1000 * 60 * 60 * 24));
        const hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
        const minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
        const seconds = Math.floor((distance % (1000 * 60)) / 1000);

        const daysEl = document.getElementById("days");
        const hoursEl = document.getElementById("hours");
        const minutesEl = document.getElementById("minutes");
        const secondsEl = document.getElementById("seconds");

        if (daysEl) daysEl.innerText = days.toString().padStart(2, '0');
        if (hoursEl) hoursEl.innerText = hours.toString().padStart(2, '0');
        if (minutesEl) minutesEl.innerText = minutes.toString().padStart(2, '0');
        if (secondsEl) secondsEl.innerText = seconds.toString().padStart(2, '0');
    }

    // Update countdown every second
    updateCountdown();
    setInterval(updateCountdown, 1000);

    // ===================
    // Mobile Menu Toggle
    // ===================
    function initMobileMenu() {
        const menuBtn = document.getElementById('mobile-menu-btn');
        const mobileMenu = document.getElementById('mobile-menu');
        
        if (menuBtn && mobileMenu) {
            menuBtn.addEventListener('click', () => {
                mobileMenu.classList.toggle('hidden');
                const icon = menuBtn.querySelector('i');
                if (icon) {
                    icon.classList.toggle('fa-bars');
                    icon.classList.toggle('fa-xmark');
                }
            });
            
            // Close menu when clicking a link
            const menuLinks = mobileMenu.querySelectorAll('a[href^="#"]');
            menuLinks.forEach(link => {
                link.addEventListener('click', () => {
                    mobileMenu.classList.add('hidden');
                    const icon = menuBtn.querySelector('i');
                    if (icon) {
                        icon.classList.add('fa-bars');
                        icon.classList.remove('fa-xmark');
                    }
                });
            });
        }
    }

    // ===================
    // Scroll Animations
    // ===================
    function initScrollAnimations() {
        const observerOptions = {
            root: null,
            rootMargin: '0px',
            threshold: 0.1
        };

        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    entry.target.classList.add('active');
                    entry.target.classList.remove('opacity-0', 'translate-y-10');
                }
            });
        }, observerOptions);

        // Observe all cards
        document.querySelectorAll('.card-shadow').forEach(card => {
            card.classList.add('opacity-0', 'translate-y-10', 'transition-all', 'duration-700');
            observer.observe(card);
        });
    }

    // ===================
    // Smooth Scroll for Navigation
    // ===================
    function initSmoothScroll() {
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function(e) {
                e.preventDefault();
                const targetId = this.getAttribute('href');
                const target = document.querySelector(targetId);
                
                if (target) {
                    const navHeight = document.querySelector('nav')?.offsetHeight || 0;
                    const targetPosition = target.getBoundingClientRect().top + window.pageYOffset - navHeight;
                    
                    window.scrollTo({
                        top: targetPosition,
                        behavior: 'smooth'
                    });
                }
            });
        });
    }

    // ===================
    // Navigation Background on Scroll
    // ===================
    function initNavScroll() {
        const nav = document.querySelector('nav');
        
        if (nav) {
            window.addEventListener('scroll', () => {
                if (window.scrollY > 100) {
                    nav.classList.add('shadow-md');
                } else {
                    nav.classList.remove('shadow-md');
                }
            });
        }
    }

    // ===================
    // Easter Egg: Ampersand Heart
    // ===================
    function initAmpersandEasterEgg() {
        const ampersand = document.getElementById('ampersand-easter-egg');
        if (!ampersand) return;

        let hoverTimer = null;
        let isHeartMode = false;
        const HOVER_DURATION = 2500; // 2.5 seconds
        const RESET_DURATION = 5000; // Reset after 5 seconds

        // Wedding color palette for confetti
        const confettiColors = [
            '#E06C75', // rose
            '#F2A93B', // marigold
            '#9D8EB5', // lavender
            '#8FA876', // leaf
            '#D97757', // clay
        ];

        function createConfetti() {
            const rect = ampersand.getBoundingClientRect();
            const centerX = rect.left + rect.width / 2;
            const centerY = rect.top + rect.height / 2;
            const particleCount = 35;

            for (let i = 0; i < particleCount; i++) {
                const particle = document.createElement('div');
                particle.className = 'confetti-particle';
                
                // Random shape
                const shapes = ['circle', 'square', 'heart'];
                particle.classList.add(shapes[Math.floor(Math.random() * shapes.length)]);
                
                // Random color
                const color = confettiColors[Math.floor(Math.random() * confettiColors.length)];
                particle.style.backgroundColor = color;
                
                // Random position around the ampersand
                const angle = (Math.PI * 2 * i) / particleCount + (Math.random() - 0.5);
                const distance = 20 + Math.random() * 60;
                const x = centerX + Math.cos(angle) * distance;
                const y = centerY + Math.sin(angle) * distance;
                
                particle.style.left = x + 'px';
                particle.style.top = y + 'px';
                
                // Random animation variation
                particle.style.animationDuration = (2 + Math.random() * 2) + 's';
                particle.style.animationDelay = (Math.random() * 0.3) + 's';
                
                // Random size
                const size = 6 + Math.random() * 8;
                particle.style.width = size + 'px';
                particle.style.height = size + 'px';
                
                document.body.appendChild(particle);
                
                // Remove particle after animation
                setTimeout(() => {
                    particle.remove();
                }, 4000);
            }
        }

        function triggerEasterEgg() {
            if (isHeartMode) return;
            
            isHeartMode = true;
            ampersand.textContent = '❤';
            ampersand.classList.add('heart-mode');
            
            // Create confetti burst
            createConfetti();
            
            // Reset after a delay
            setTimeout(() => {
                ampersand.textContent = '&';
                ampersand.classList.remove('heart-mode');
                isHeartMode = false;
            }, RESET_DURATION);
        }

        ampersand.addEventListener('mouseenter', () => {
            if (isHeartMode) return;
            hoverTimer = setTimeout(triggerEasterEgg, HOVER_DURATION);
        });

        ampersand.addEventListener('mouseleave', () => {
            if (hoverTimer) {
                clearTimeout(hoverTimer);
                hoverTimer = null;
            }
        });

        // Touch support for mobile (long press)
        let touchTimer = null;
        ampersand.addEventListener('touchstart', (e) => {
            if (isHeartMode) return;
            e.preventDefault(); // Prevent context menu
            touchTimer = setTimeout(triggerEasterEgg, HOVER_DURATION);
        });

        ampersand.addEventListener('touchend', (e) => {
            e.preventDefault();
            if (touchTimer) {
                clearTimeout(touchTimer);
                touchTimer = null;
            }
        });

        ampersand.addEventListener('touchcancel', () => {
            if (touchTimer) {
                clearTimeout(touchTimer);
                touchTimer = null;
            }
        });

        // Prevent context menu on long press
        ampersand.addEventListener('contextmenu', (e) => {
            e.preventDefault();
        });
    }

    // ===================
    // Relative Time for Last Updated
    // ===================
    function initLastUpdated() {
        const el = document.getElementById('last-updated');
        if (!el) return;

        const timestamp = parseInt(el.dataset.timestamp, 10) * 1000;
        const label = el.dataset.label || 'Last updated';
        const now = Date.now();
        const diff = now - timestamp;

        // Calculate relative time
        const seconds = Math.floor(diff / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        let relativeTime;
        if (days > 30) {
            const date = new Date(timestamp);
            relativeTime = date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
        } else if (days > 0) {
            relativeTime = days === 1 ? '1 day ago' : `${days} days ago`;
        } else if (hours > 0) {
            relativeTime = hours === 1 ? '1 hour ago' : `${hours} hours ago`;
        } else if (minutes > 0) {
            relativeTime = minutes === 1 ? '1 minute ago' : `${minutes} minutes ago`;
        } else {
            relativeTime = 'just now';
        }

        el.textContent = `${label} ${relativeTime}`;
    }

    // ===================
    // FAQ Accordion
    // ===================
    function initFAQ() {
        document.querySelectorAll('.faq-trigger').forEach(trigger => {
            trigger.addEventListener('click', () => {
                const content = trigger.nextElementSibling;
                const icon = trigger.querySelector('.fa-plus, .fa-minus');
                const isExpanded = content.style.maxHeight;

                // Close all others
                document.querySelectorAll('.faq-content').forEach(c => {
                    if (c !== content) {
                        c.style.maxHeight = null;
                    }
                });
                document.querySelectorAll('.faq-trigger .fa-minus').forEach(i => {
                    if (i !== icon) {
                        i.classList.remove('fa-minus');
                        i.classList.add('fa-plus');
                    }
                });

                if (isExpanded) {
                    content.style.maxHeight = null;
                    icon.classList.remove('fa-minus');
                    icon.classList.add('fa-plus');
                } else {
                    content.style.maxHeight = content.scrollHeight + "px";
                    icon.classList.remove('fa-plus');
                    icon.classList.add('fa-minus');
                }
            });
        });
    }

    // ===================
    // RSVP Form
    // ===================
    function initRSVP() {
        const rsvpSection = document.getElementById('rsvp');
        const rsvpCard = document.getElementById('rsvp-card');
        if (!rsvpSection || !rsvpCard) return;

        const params = new URLSearchParams(window.location.search);
        const code = params.get('code');
        if (!code) return;

        let apiBase = rsvpCard.dataset.apiBase || '';
        if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
            apiBase = 'http://localhost:8081/api/v1';
        }
        const endpoint = `${apiBase}/invite/${encodeURIComponent(code)}`;

        const loadingEl = document.getElementById('rsvp-loading');
        const errorEl = document.getElementById('rsvp-error');
        const errorMessageEl = document.getElementById('rsvp-error-message');
        const formWrapper = document.getElementById('rsvp-form-wrapper');
        const thanksEl = document.getElementById('rsvp-thanks');
        const thanksNameEl = document.getElementById('rsvp-thanks-name');
        const guestNameEl = document.getElementById('rsvp-guest-name');
        const plusOneField = document.getElementById('rsvp-plus-one-field');
        const kidsSelect = document.getElementById('rsvp-kids');
        const kidsField = document.getElementById('rsvp-kids-field');
        const kidsMaxEl = document.getElementById('rsvp-kids-max');
        const form = document.getElementById('rsvp-form');
        const submitBtn = document.getElementById('rsvp-submit');
        const dietaryInput = document.getElementById('rsvp-dietary');
        const messageInput = document.getElementById('rsvp-message');
        const songInput = document.getElementById('rsvp-song');

        if (!loadingEl || !errorEl || !errorMessageEl || !formWrapper || !thanksEl || !guestNameEl || !form || !submitBtn) {
            return;
        }

        const errorMissingPlusOne = rsvpCard.dataset.errorMissingPlusOne || 'Please indicate if you will bring a +1.';
        const errorMissingKids = rsvpCard.dataset.errorMissingKids || 'Please select the number of kids.';
        const errorGeneric = rsvpCard.dataset.errorGeneric || 'Something went wrong. Please try again.';

        function setLoading(isLoading) {
            if (loadingEl) loadingEl.classList.toggle('hidden', !isLoading);
        }

        function showError(message) {
            if (!errorEl || !errorMessageEl) return;
            errorMessageEl.textContent = message;
            errorEl.classList.remove('hidden');
        }

        function clearError() {
            if (!errorEl || !errorMessageEl) return;
            errorMessageEl.textContent = '';
            errorEl.classList.add('hidden');
        }

        function resetSubmit(isSubmitting) {
            const defaultText = submitBtn.dataset.defaultText || submitBtn.textContent;
            const loadingText = submitBtn.dataset.loadingText || defaultText;
            submitBtn.disabled = isSubmitting;
            submitBtn.textContent = isSubmitting ? loadingText : defaultText;
            submitBtn.classList.toggle('opacity-70', isSubmitting);
        }

        function populateSelect(select, maxValue, startAtZero) {
            if (!select) return;
            select.innerHTML = '';
            const placeholder = document.createElement('option');
            placeholder.value = '';
            placeholder.disabled = true;
            placeholder.selected = true;
            placeholder.textContent = select.getAttribute('data-placeholder') || select.querySelector('option')?.textContent || '';
            select.appendChild(placeholder);

            const start = startAtZero ? 0 : 1;
            for (let i = start; i <= maxValue; i++) {
                const option = document.createElement('option');
                option.value = String(i);
                option.textContent = String(i);
                select.appendChild(option);
            }
        }

        async function parseErrorResponse(response) {
            const text = await response.text();
            if (!text) return null;
            try {
                return JSON.parse(text);
            } catch (err) {
                return text;
            }
        }

        function formatErrorMessage(payload) {
            if (!payload) return errorGeneric;
            if (typeof payload === 'string') return payload;
            if (typeof payload === 'object') {
                if (payload.error) return payload.error;
                return JSON.stringify(payload, null, 2);
            }
            return String(payload);
        }

        async function loadInvite() {
            rsvpSection.classList.remove('hidden');
            setLoading(true);
            clearError();

            try {
                const response = await fetch(endpoint, { headers: { 'Accept': 'application/json' } });
                if (!response.ok) {
                    const payload = await parseErrorResponse(response);
                    throw new Error(formatErrorMessage(payload));
                }

                const data = await response.json();
                setLoading(false);
                clearError();

                const name = data?.name || '';
                if (guestNameEl) guestNameEl.textContent = name;
                if (thanksNameEl) thanksNameEl.textContent = name;

                const maxAdults = Number(data?.max_adults || 0);
                const maxKids = Number(data?.max_kids || 0);

                if (kidsMaxEl) kidsMaxEl.textContent = maxKids ? maxKids : '';

                // Update invite size message
                const totalPeople = maxAdults + maxKids;
                const inviteSizeText = rsvpCard.dataset.inviteSizeText || '';
                const inviteSizeEl = document.getElementById('rsvp-invite-size');
                
                if (inviteSizeEl && inviteSizeText) {
                    inviteSizeEl.textContent = inviteSizeText.replace('{n}', totalPeople);
                }

                // Show +1 field only if max_adults is 2
                const hasPlusOne = maxAdults === 2;
                if (plusOneField) {
                    if (!hasPlusOne) {
                        plusOneField.classList.add('hidden');
                    } else {
                        plusOneField.classList.remove('hidden');
                    }
                }

                if (kidsSelect) {
                    populateSelect(kidsSelect, maxKids, true);
                }

                if (kidsField) {
                    if (maxKids <= 0) {
                        kidsField.classList.add('hidden');
                    } else {
                        kidsField.classList.remove('hidden');
                    }
                }

                if (data?.has_responded) {
                    thanksEl.classList.remove('hidden');
                    formWrapper.classList.add('hidden');
                } else {
                    formWrapper.classList.remove('hidden');
                    thanksEl.classList.add('hidden');
                }

                form.addEventListener('submit', async (event) => {
                    event.preventDefault();
                    clearError();

                    // Validate kids field if applicable
                    if (maxKids > 0 && !kidsSelect?.value) {
                        showError(errorMissingKids);
                        return;
                    }

                    const payload = {
                        attending: true,  // Always true when form is submitted
                        dietary_info: dietaryInput?.value?.trim() || '',
                        message_for_us: messageInput?.value?.trim() || '',
                        song_request: songInput?.value?.trim() || ''
                    };

                    // Determine adult_count based on +1 checkbox
                    if (maxAdults === 2) {
                        const plusOneCheckbox = document.getElementById('rsvp-plus-one-checkbox');
                        payload.adult_count = plusOneCheckbox?.checked ? 2 : 1;
                    } else {
                        // If max_adults is 1, always send 1
                        payload.adult_count = 1;
                    }
                    
                    if (maxKids > 0) {
                        payload.kid_count = Number(kidsSelect.value || 0);
                    }

                    try {
                        resetSubmit(true);
                        const response = await fetch(`${endpoint}/rsvp`, {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                                'Accept': 'application/json'
                            },
                            body: JSON.stringify(payload)
                        });

                        if (!response.ok) {
                            const payloadError = await parseErrorResponse(response);
                            throw new Error(formatErrorMessage(payloadError));
                        }

                        formWrapper.classList.add('hidden');
                        thanksEl.classList.remove('hidden');
                    } catch (err) {
                        showError(err?.message || errorGeneric);
                    } finally {
                        resetSubmit(false);
                    }
                });
            } catch (err) {
                setLoading(false);
                showError(err?.message || errorGeneric);
            }
        }

        loadInvite();
    }

    // ===================
    // Place Cards Modal
    // ===================
    function initPlaceCards() {
        // Navigate to a specific place modal
        function navigateToPlace(group, index) {
            const currentModal = document.querySelector('.place-modal.active');
            const targetModal = document.getElementById(`place-modal-${group}-${index}`);
            if (targetModal) {
                // Add active to new modal BEFORE removing from old to prevent backdrop flicker
                targetModal.classList.add('active');
                if (currentModal && currentModal !== targetModal) {
                    currentModal.classList.remove('active');
                }
                // Reset scroll position in new modal
                const content = targetModal.querySelector('.flex-grow.overflow-y-auto');
                if (content) content.scrollTop = 0;
            }
        }

        // Open modal when clicking "Learn more" button
        document.querySelectorAll('.place-card-trigger').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();
                const modal = document.getElementById(btn.dataset.target);
                if (modal) {
                    modal.classList.add('active');
                    document.body.classList.add('place-modal-open');
                }
            });
        });

        // Close modal on backdrop click
        const sharedBackdrop = document.getElementById('place-modal-backdrop');
        if (sharedBackdrop) {
            sharedBackdrop.addEventListener('click', () => {
                const activeModal = document.querySelector('.place-modal.active');
                if (activeModal) {
                    activeModal.classList.remove('active');
                    document.body.classList.remove('place-modal-open');
                }
            });
        }

        // Close modal on close button click
        document.querySelectorAll('.place-modal-close').forEach(btn => {
            btn.addEventListener('click', () => {
                const modal = btn.closest('.place-modal');
                if (modal) {
                    modal.classList.remove('active');
                    document.body.classList.remove('place-modal-open');
                }
            });
        });

        // Navigate to previous place
        document.querySelectorAll('.place-modal-prev').forEach(btn => {
            btn.addEventListener('click', () => {
                if (btn.disabled) return;
                const group = btn.dataset.group;
                const index = parseInt(btn.dataset.index, 10);
                navigateToPlace(group, index - 1);
            });
        });

        // Navigate to next place
        document.querySelectorAll('.place-modal-next').forEach(btn => {
            btn.addEventListener('click', () => {
                if (btn.disabled) return;
                const group = btn.dataset.group;
                const index = parseInt(btn.dataset.index, 10);
                navigateToPlace(group, index + 1);
            });
        });

        // Keyboard navigation: Escape to close, Arrow keys to navigate
        document.addEventListener('keydown', (e) => {
            const activeModal = document.querySelector('.place-modal.active');
            if (!activeModal) return;

            if (e.key === 'Escape') {
                activeModal.classList.remove('active');
                document.body.classList.remove('place-modal-open');
            } else if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
                const navBtn = activeModal.querySelector(
                    e.key === 'ArrowLeft' ? '.place-modal-prev' : '.place-modal-next'
                );
                if (navBtn && !navBtn.disabled) {
                    navBtn.click();
                }
            }
        });
    }

    // ===================
    // Initialize Everything
    // ===================
    document.addEventListener('DOMContentLoaded', () => {
        initMobileMenu();
        initScrollAnimations();
        initSmoothScroll();
        initNavScroll();
        initAmpersandEasterEgg();
        initLastUpdated();
        initPlaceCards();
        initFAQ();
        initRSVP();
    });

})();
