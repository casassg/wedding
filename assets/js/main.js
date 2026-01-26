/**
 * Laura & Gerard Wedding - Alpine.js Components
 * Declarative, reactive DOM handling with Alpine.js
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
    // Confetti Effect (shared utility)
    // ===================
    const confettiColors = [
        '#E06C75', // rose
        '#F2A93B', // marigold
        '#9D8EB5', // lavender
        '#8FA876', // leaf
        '#D97757', // clay
    ];

    window.createConfetti = function(targetElement = null, particleCount = 35, explosive = false) {
        let centerX, centerY;
        
        if (targetElement) {
            const rect = targetElement.getBoundingClientRect();
            centerX = rect.left + rect.width / 2;
            centerY = rect.top + rect.height / 2;
        } else {
            centerX = window.innerWidth / 2;
            centerY = window.innerHeight / 2;
        }

        for (let i = 0; i < particleCount; i++) {
            const particle = document.createElement('div');
            particle.className = 'confetti-particle';
            
            const shapes = ['circle', 'square', 'heart'];
            particle.classList.add(shapes[Math.floor(Math.random() * shapes.length)]);
            
            const color = confettiColors[Math.floor(Math.random() * confettiColors.length)];
            particle.style.backgroundColor = color;
            
            let x, y;
            if (explosive) {
                const angle = (Math.PI * 2 * i) / particleCount + (Math.random() - 0.5) * 0.8;
                const distance = 100 + Math.random() * Math.max(window.innerWidth, window.innerHeight) * 0.8;
                x = centerX + Math.cos(angle) * distance;
                y = centerY + Math.sin(angle) * distance;
            } else {
                const angle = (Math.PI * 2 * i) / particleCount + (Math.random() - 0.5);
                const distance = 20 + Math.random() * 60;
                x = centerX + Math.cos(angle) * distance;
                y = centerY + Math.sin(angle) * distance;
            }
            
            particle.style.left = x + 'px';
            particle.style.top = y + 'px';
            
            const duration = explosive ? (3 + Math.random() * 3) : (2 + Math.random() * 2);
            particle.style.animationDuration = duration + 's';
            particle.style.animationDelay = (Math.random() * 0.3) + 's';
            
            const baseSize = explosive ? 8 : 6;
            const sizeVariation = explosive ? 12 : 8;
            const size = baseSize + Math.random() * sizeVariation;
            particle.style.width = size + 'px';
            particle.style.height = size + 'px';
            
            document.body.appendChild(particle);
            
            const removeDelay = explosive ? 6000 : 4000;
            setTimeout(() => {
                particle.remove();
            }, removeDelay);
        }
    };

    // ===================
    // Smooth Scroll Utility
    // ===================
    window.smoothScrollTo = function(targetId) {
        const target = document.querySelector(targetId);
        if (!target) return;
        
        const navHeight = document.querySelector('nav')?.offsetHeight || 0;
        const targetPosition = target.getBoundingClientRect().top + window.pageYOffset - navHeight;
        
        window.scrollTo({
            top: targetPosition,
            behavior: 'smooth'
        });
    };

    // ===================
    // Alpine.js Component Definitions
    // ===================
    document.addEventListener('alpine:init', () => {
        
        // -----------------------
        // Navigation Component (mobile menu + scroll shadow)
        // -----------------------
        Alpine.data('navigation', () => ({
            menuOpen: false,
            scrolled: false,
            
            init() {
                this.checkScroll();
            },
            
            toggleMenu() {
                this.menuOpen = !this.menuOpen;
            },
            
            closeMenu() {
                this.menuOpen = false;
            },
            
            checkScroll() {
                this.scrolled = window.scrollY > 100;
            }
        }));

        // -----------------------
        // Countdown Timer Component
        // -----------------------
        Alpine.data('countdown', () => ({
            days: '00',
            hours: '00',
            minutes: '00',
            seconds: '00',
            ended: false,
            weddingDate: new Date("2026-12-19T16:00:00-06:00").getTime(),
            interval: null,
            
            init() {
                this.updateCountdown();
                this.interval = setInterval(() => this.updateCountdown(), 1000);
            },
            
            destroy() {
                if (this.interval) clearInterval(this.interval);
            },
            
            updateCountdown() {
                const now = new Date().getTime();
                const distance = this.weddingDate - now;
                
                if (distance < 0) {
                    this.ended = true;
                    if (this.interval) clearInterval(this.interval);
                    return;
                }
                
                this.days = Math.floor(distance / (1000 * 60 * 60 * 24)).toString().padStart(2, '0');
                this.hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)).toString().padStart(2, '0');
                this.minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60)).toString().padStart(2, '0');
                this.seconds = Math.floor((distance % (1000 * 60)) / 1000).toString().padStart(2, '0');
            }
        }));

        // -----------------------
        // FAQ Accordion Store (tracks which FAQ is open - only one at a time)
        // -----------------------
        Alpine.store('faq', {
            openId: null,
            
            toggle(id) {
                this.openId = this.openId === id ? null : id;
            },
            
            isOpen(id) {
                return this.openId === id;
            }
        });

        // -----------------------
        // FAQ Accordion Item Component
        // -----------------------
        Alpine.data('faqItem', (id) => ({
            id: id,
            
            get open() {
                return Alpine.store('faq').isOpen(this.id);
            },
            
            toggle() {
                Alpine.store('faq').toggle(this.id);
            }
        }));

        // -----------------------
        // Ampersand Easter Egg Component
        // -----------------------
        Alpine.data('ampersandEgg', () => ({
            isHeartMode: false,
            hoverTimer: null,
            HOVER_DURATION: 2500,
            RESET_DURATION: 5000,
            
            startHover() {
                if (this.isHeartMode) return;
                this.hoverTimer = setTimeout(() => this.trigger(), this.HOVER_DURATION);
            },
            
            endHover() {
                if (this.hoverTimer) {
                    clearTimeout(this.hoverTimer);
                    this.hoverTimer = null;
                }
            },
            
            trigger() {
                if (this.isHeartMode) return;
                
                this.isHeartMode = true;
                window.createConfetti(this.$el);
                
                setTimeout(() => {
                    this.isHeartMode = false;
                }, this.RESET_DURATION);
            }
        }));

        // -----------------------
        // Last Updated Component
        // -----------------------
        Alpine.data('lastUpdated', () => ({
            relativeTime: '',
            
            init() {
                const timestamp = parseInt(this.$el.dataset.timestamp, 10) * 1000;
                const label = this.$el.dataset.label || 'Last updated';
                
                const now = Date.now();
                const diff = now - timestamp;
                
                const seconds = Math.floor(diff / 1000);
                const minutes = Math.floor(seconds / 60);
                const hours = Math.floor(minutes / 60);
                const days = Math.floor(hours / 24);
                
                let timeStr;
                if (days > 30) {
                    const date = new Date(timestamp);
                    timeStr = date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
                } else if (days > 0) {
                    timeStr = days === 1 ? '1 day ago' : `${days} days ago`;
                } else if (hours > 0) {
                    timeStr = hours === 1 ? '1 hour ago' : `${hours} hours ago`;
                } else if (minutes > 0) {
                    timeStr = minutes === 1 ? '1 minute ago' : `${minutes} minutes ago`;
                } else {
                    timeStr = 'just now';
                }
                
                this.relativeTime = `${label} ${timeStr}`;
            }
        }));

        // -----------------------
        // Place Modals Store (global state for modals)
        // -----------------------
        Alpine.store('placeModals', {
            activeModal: null,
            _group: null,
            _index: null,
            _total: null,
            
            open(modalId, group, index, total) {
                this.activeModal = modalId;
                this._group = group;
                this._index = index;
                this._total = total;
                document.body.classList.add('place-modal-open');
            },
            
            close() {
                this.activeModal = null;
                this._group = null;
                this._index = null;
                this._total = null;
                document.body.classList.remove('place-modal-open');
            },
            
            navigate(newIndex) {
                if (this._group === null || newIndex < 0 || newIndex >= this._total) return;
                this._index = newIndex;
                this.activeModal = `place-modal-${this._group}-${newIndex}`;
                // Reset scroll position
                requestAnimationFrame(() => {
                    const content = document.querySelector('.place-modal.active .flex-grow.overflow-y-auto');
                    if (content) content.scrollTop = 0;
                });
            },
            
            prev() {
                if (this._index > 0) {
                    this.navigate(this._index - 1);
                }
            },
            
            next() {
                if (this._index < this._total - 1) {
                    this.navigate(this._index + 1);
                }
            },
            
            isActive(modalId) {
                return this.activeModal === modalId;
            },
            
            get hasPrev() {
                return this._index !== null && this._index > 0;
            },
            
            get hasNext() {
                return this._index !== null && this._total !== null && this._index < this._total - 1;
            }
        });

        // -----------------------
        // Place Modal Component (individual modal instance)
        // -----------------------
        Alpine.data('placeModal', (group, index, total) => ({
            group: group,
            index: index,
            total: total,
            
            get isOpen() {
                return Alpine.store('placeModals').isActive(`place-modal-${this.group}-${this.index}`);
            },
            
            get hasPrev() {
                return Alpine.store('placeModals').hasPrev;
            },
            
            get hasNext() {
                return Alpine.store('placeModals').hasNext;
            },
            
            open() {
                Alpine.store('placeModals').open(
                    `place-modal-${this.group}-${this.index}`,
                    this.group,
                    this.index,
                    this.total
                );
            },
            
            close() {
                Alpine.store('placeModals').close();
            },
            
            prev() {
                Alpine.store('placeModals').prev();
            },
            
            next() {
                Alpine.store('placeModals').next();
            }
        }));

        // -----------------------
        // RSVP Form Component
        // -----------------------
        Alpine.data('rsvpForm', () => ({
            loading: true,
            error: null,
            invite: null,
            submitted: false,
            submitting: false,
            code: null,
            
            formData: {
                plusOne: false,
                kidCount: '',
                dietaryInfo: '',
                message: '',
                song: ''
            },
            
            // Error messages from data attributes
            errorMissingKids: '',
            errorGeneric: '',
            inviteSizeText: '',
            
            init() {
                // Read config from data attributes
                const el = this.$el;
                this.errorMissingKids = el.dataset.errorMissingKids || 'Please select the number of kids.';
                this.errorGeneric = el.dataset.errorGeneric || 'Something went wrong. Please try again.';
                this.inviteSizeText = el.dataset.inviteSizeText || '';
                
                // Check for invite code
                const params = new URLSearchParams(window.location.search);
                this.code = params.get('code');
                
                if (this.code) {
                    this.loadInvite();
                } else {
                    this.loading = false;
                }
            },
            
            get apiBase() {
                if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
                    return 'http://localhost:8080/api/v1';
                }
                return this.$el.dataset.apiBase || '';
            },
            
            get showSection() {
                return this.code !== null;
            },
            
            get showPlusOne() {
                return this.invite?.max_adults === 2;
            },
            
            get showKids() {
                return (this.invite?.max_kids || 0) > 0;
            },
            
            get kidsOptions() {
                if (!this.invite) return [];
                const max = this.invite.max_kids || 0;
                return Array.from({ length: max + 1 }, (_, i) => i);
            },
            
            get inviteSizeMessage() {
                if (!this.invite || !this.inviteSizeText) return '';
                const total = (this.invite.max_adults || 0) + (this.invite.max_kids || 0);
                return this.inviteSizeText.replace('{n}', total);
            },
            
            async loadInvite() {
                this.loading = true;
                this.error = null;
                
                try {
                    const response = await fetch(`${this.apiBase}/invite/${encodeURIComponent(this.code)}`, {
                        headers: { 'Accept': 'application/json' }
                    });
                    
                    if (!response.ok) {
                        const payload = await this.parseErrorResponse(response);
                        throw new Error(this.formatErrorMessage(payload));
                    }
                    
                    this.invite = await response.json();
                    this.submitted = this.invite.has_responded;
                } catch (err) {
                    this.error = err.message || this.errorGeneric;
                } finally {
                    this.loading = false;
                }
            },
            
            async submitRSVP() {
                if (this.submitting) return;
                
                this.error = null;
                
                // Validate kids field if applicable
                if (this.showKids && this.formData.kidCount === '') {
                    this.error = this.errorMissingKids;
                    return;
                }
                
                const payload = {
                    dietary_info: this.formData.dietaryInfo.trim(),
                    message_for_us: this.formData.message.trim(),
                    song_request: this.formData.song.trim()
                };
                
                // Determine adult_count based on +1 checkbox
                if (this.invite.max_adults === 2) {
                    payload.adult_count = this.formData.plusOne ? 2 : 1;
                } else {
                    payload.adult_count = 1;
                }
                
                if (this.showKids) {
                    payload.kid_count = parseInt(this.formData.kidCount) || 0;
                }
                
                try {
                    this.submitting = true;
                    
                    const response = await fetch(`${this.apiBase}/invite/${encodeURIComponent(this.code)}/rsvp`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'Accept': 'application/json'
                        },
                        body: JSON.stringify(payload)
                    });
                    
                    if (!response.ok) {
                        const payloadError = await this.parseErrorResponse(response);
                        throw new Error(this.formatErrorMessage(payloadError));
                    }
                    
                    this.submitted = true;
                    
                    // Trigger confetti explosion
                    setTimeout(() => {
                        window.createConfetti(this.$el, 100, true);
                    }, 200);
                } catch (err) {
                    this.error = err.message || this.errorGeneric;
                } finally {
                    this.submitting = false;
                }
            },
            
            async parseErrorResponse(response) {
                const text = await response.text();
                if (!text) return null;
                try {
                    return JSON.parse(text);
                } catch (err) {
                    return text;
                }
            },
            
            formatErrorMessage(payload) {
                if (!payload) return this.errorGeneric;
                if (typeof payload === 'string') return payload;
                if (typeof payload === 'object') {
                    if (payload.error) return payload.error;
                    return JSON.stringify(payload, null, 2);
                }
                return String(payload);
            }
        }));

        // -----------------------
        // Scroll Animation Component (IntersectionObserver)
        // -----------------------
        Alpine.data('scrollReveal', () => ({
            visible: false,
            
            init() {
                const observer = new IntersectionObserver((entries) => {
                    entries.forEach(entry => {
                        if (entry.isIntersecting) {
                            this.visible = true;
                            observer.unobserve(entry.target);
                        }
                    });
                }, {
                    root: null,
                    rootMargin: '0px',
                    threshold: 0.1
                });
                
                observer.observe(this.$el);
            }
        }));
    });

    // ===================
    // Initialize scroll reveal after Alpine is ready
    // ===================
    document.addEventListener('alpine:initialized', () => {
        // Auto-apply scroll reveal to .card-shadow and .card-glass elements
        // Run after Alpine so we don't conflict with Alpine-managed elements
        document.querySelectorAll('.card-shadow, .card-glass').forEach(card => {
            // Skip elements that are inside Alpine-controlled sections that handle their own visibility
            if (card.closest('[x-show]') || card.closest('[x-if]')) return;
            
            // Add initial hidden state
            card.classList.add('opacity-0', 'translate-y-10', 'transition-all', 'duration-700');
            
            const observer = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        entry.target.classList.remove('opacity-0', 'translate-y-10');
                        observer.unobserve(entry.target);
                    }
                });
            }, { threshold: 0.1 });
            
            observer.observe(card);
        });
    });

})();
