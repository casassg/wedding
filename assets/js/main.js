/**
 * Laura & Gerard Wedding - Alpine.js Components
 * Declarative, reactive DOM handling with Alpine.js
 */

(function() {
    'use strict';

    function apiBaseUrl() {
        const host = window.location.hostname;
        if (host === 'localhost' || host === '127.0.0.1') {
            return 'http://localhost:8080/api/v1';
        }
        if (host === 'lauraygerard.wedding' || host === 'www.lauraygerard.wedding') {
            return 'https://api.lauraygerard.wedding/api/v1';
        }
        return `${window.location.origin}/api/v1`;
    }

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

        // Only redirect on the root paths - if user navigated to /es/ or /ca/ explicitly, respect that
        const isRootPath = path === '/';
        
        if (!isRootPath) {
            // User explicitly navigated to a language-specific path, mark it as selected
            localStorage.setItem('lang-selected', 'true');
            return;
        }

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
            
            // Mark that we've redirected based on browser language
            localStorage.setItem('lang-selected', 'true');
            
            window.location.replace(newUrl);
        }
    }

    // Run language detection immediately
    initLanguageDetection();

    // ===================
    // Preserve Code Parameter on Internal Links
    // ===================
    function preserveCodeParameter() {
        const code = new URLSearchParams(window.location.search).get('code');
        if (!code) return;

        // Add click handler to all internal links
        document.addEventListener('click', function(e) {
            const link = e.target.closest('a');
            if (!link) return;

            const href = link.getAttribute('href');
            if (!href) return;

            // Check if it's an internal relative link (starts with / or doesn't have a protocol)
            const isRelative = href.startsWith('/') || (!href.startsWith('http://') && !href.startsWith('https://') && !href.startsWith('#') && !href.startsWith('mailto:') && !href.startsWith('tel:'));
            
            if (isRelative) {
                // Prevent default navigation
                e.preventDefault();
                
                // Build new URL with code parameter
                const url = new URL(link.href, window.location.origin);
                url.searchParams.set('code', code);
                
                // Navigate to the new URL
                window.location.href = url.toString();
            }
        });
    }

    // Run after DOM is loaded
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', preserveCodeParameter);
    } else {
        preserveCodeParameter();
    }

    // ===================
    // Dynamic Stripe Link (EUR for EU visitors, USD otherwise)
    // ===================
    function localizeStripeLinks() {
        const tz = Intl.DateTimeFormat().resolvedOptions().timeZone || '';
        if (!tz.startsWith('Europe/')) return;

        const path = window.location.pathname;
        const locale = path.includes('/es/') ? 'es-419' :
                       path.includes('/ca/') ? 'es' : 'en';

        const eurBase = 'https://donate.stripe.com/dRm00igIzfOB2YMbn90Ny01';
        document.querySelectorAll('a[href*="donate.stripe.com"]').forEach(link => {
            link.href = eurBase + '?locale=' + locale;
        });
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', localizeStripeLinks);
    } else {
        localizeStripeLinks();
    }

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
            
            // Attendance state: null = undecided (main form), false = declining
            declining: false,
            // Whether the guest confirmed attending (for thank-you variant)
            confirmedAttending: true,
            
            // Envelope state
            envelopeOpened: false,
            envelopeAnimating: false,
            
            formData: {
                plusOne: '',
                kidCount: '',
                dietaryInfo: '',
                message: '',
                song: ''
            },
            
            // Localized messages from data attributes
            errorMissingPlusOne: '',
            errorMissingKids: '',
            errorMissingFields: '',
            errorGeneric: '',
            inviteSizeOne: '',
            inviteSizeOther: '',
            fillInfo: '',
            fillInfoPlural: '',
            
            init() {
                // Read config from data attributes
                const el = this.$el;
                this.errorMissingPlusOne = el.dataset.errorMissingPlusOne || 'Please tell us how many of you are coming.';
                this.errorMissingKids = el.dataset.errorMissingKids || 'Please select the number of kids.';
                this.errorMissingFields = el.dataset.errorMissingFields || 'Please fill in all the fields.';
                this.errorGeneric = el.dataset.errorGeneric || 'Something went wrong. Please try again.';
                this.inviteSizeOne = el.dataset.inviteSizeOne || '';
                this.inviteSizeOther = el.dataset.inviteSizeOther || '';
                this.fillInfo = el.dataset.fillInfo || '';
                this.fillInfoPlural = el.dataset.fillInfoPlural || '';
                
                // Check for invite code
                const params = new URLSearchParams(window.location.search);
                this.code = params.get('code');
                
                // Restore envelope state from localStorage (skip on localhost)
                if (this.code && !this.isLocalhost) {
                    const stored = localStorage.getItem(`envelope-opened-${this.code}`);
                    if (stored === 'true') {
                        this.envelopeOpened = true;
                    }
                }
                
                if (this.code) {
                    this.loadInvite();
                } else {
                    this.loading = false;
                }
            },
            
            get isLocalhost() {
                return window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
            },
            
            get apiBase() {
                return apiBaseUrl();
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
                if (!this.invite) return '';
                const total = (this.invite.max_adults || 0) + (this.invite.max_kids || 0);
                const template = total === 1 ? this.inviteSizeOne : this.inviteSizeOther;
                return template.replace('{n}', total);
            },
            
            get fillInfoMessage() {
                if (!this.invite) return '';
                const adults = this.invite.max_adults || 0;
                return adults > 1 ? this.fillInfoPlural : this.fillInfo;
            },

            get isRSVPComplete() {
                return Boolean((!this.showPlusOne || this.formData.plusOne !== '')
                    && (!this.showKids || this.formData.kidCount !== '')
                    && this.formData.dietaryInfo.trim()
                    && this.formData.song.trim()
                    && this.formData.message.trim());
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
                    this.confirmedAttending = this.invite.is_attending;
                    
                    // Skip envelope animation if already responded
                    if (this.submitted) {
                        this.envelopeOpened = true;
                    }
                } catch (err) {
                    this.error = err.message || this.errorGeneric;
                } finally {
                    this.loading = false;
                }
            },
            
            async submitRSVP() {
                if (this.submitting) return;
                
                this.error = null;
                
                if (this.showPlusOne && this.formData.plusOne === '') {
                    this.error = this.errorMissingPlusOne;
                    return;
                }

                if (this.showKids && this.formData.kidCount === '') {
                    this.error = this.errorMissingKids;
                    return;
                }

                if (!this.formData.dietaryInfo.trim() || !this.formData.song.trim() || !this.formData.message.trim()) {
                    this.error = this.errorMissingFields;
                    return;
                }
                
                const payload = {
                    dietary_info: this.formData.dietaryInfo.trim(),
                    message_for_us: this.formData.message.trim(),
                    song_request: this.formData.song.trim()
                };
                
                // Determine adult_count from the explicit adult-count choice.
                if (this.invite.max_adults === 2) {
                    payload.adult_count = this.formData.plusOne === 'yes' ? 2 : 1;
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
                    this.confirmedAttending = true;
                    
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
            
            startDecline() {
                this.declining = true;
                this.error = null;
            },
            
            cancelDecline() {
                this.declining = false;
                this.error = null;
            },
            
            async submitDecline() {
                if (this.submitting) return;
                
                this.error = null;
                
                const payload = {
                    adult_count: 0,
                    kid_count: 0,
                    dietary_info: '',
                    message_for_us: this.formData.message.trim(),
                    song_request: ''
                };
                
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
                    this.confirmedAttending = false;
                } catch (err) {
                    this.error = err.message || this.errorGeneric;
                } finally {
                    this.submitting = false;
                }
            },
            
            openEnvelope() {
                if (this.envelopeAnimating || this.envelopeOpened) return;
                
                this.envelopeAnimating = true;
                
                // Trigger confetti explosion from wax seal
                const sealElement = this.$el.querySelector('.wax-seal');
                window.createConfetti(sealElement || this.$el, 100, true);
                
                // Wait for animations: seal(300ms) + flap(600ms) + fade(400ms) with overlap
                setTimeout(() => {
                    this.envelopeOpened = true;
                    this.envelopeAnimating = false;
                    
                    // Store in localStorage (skip on localhost)
                    if (this.code && !this.isLocalhost) {
                        localStorage.setItem(`envelope-opened-${this.code}`, 'true');
                    }
                }, 900);
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
        // Travel Form Component
        // -----------------------
        Alpine.data('travelForm', () => ({
            // Flight data from embedded JSON (normalized on load)
            allFlights: [],
            allDepartures: [],
            hotels: [],

            // UI state
            saveStatus: '', // '', 'saving', 'saved', 'retry'
            saveTimer: null,
            saveInFlight: false,
            saveQueued: false,
            saveVersion: 0,
            flightOpen: false,
            returnFlightOpen: false,
            lang: 'en',
            inHonduras: false,

            // Two-step return bus UI state (not persisted directly; derived to/from travel.busreturn)
            returnTime: '', // 'morning' | 'afternoon' | 'none' | ''
            returnDest: '', // 'san_pedro' | 'sap' | ''

            // Bus date ISO strings derived from the configured wedding date
            thursdayDate: '',
            fridayDate: '',
            sundayDate: '',

            // The invite code — pulled from URL
            get code() {
                return new URLSearchParams(window.location.search).get('code');
            },

            // Travel form state
            travel: {
                busto: '',       // 'thursday' | 'friday' | 'none' | ''
                pickup: '',      // 'sap' | 'welchez' | ''
                flightInput: '', // free text / autocomplete input
                busreturn: '',   // 'sunday_morning_sap' | 'sunday_morning_san_pedro' | 'sunday_afternoon_san_pedro' | 'none' | ''
                hotel: '',       // hotel id | '__other__' | ''
                hotelOther: '',  // free text when hotel === '__other__'
                notes: '',       // textarea
                cocktail: '',    // 'yes' | 'no' | ''
                brunch: '',      // 'yes' | 'no' | ''
                returnDetail: '', // departure flight label (returnDest=sap) or drop-off text (returnDest=san_pedro)
            },

            // Only flights that land on the bus day at or before that day's landing
            // cutoff. Thursday's bus leaves Welchez Café at 3 PM (cutoff 1:30 PM);
            // Friday's leaves an hour earlier, at 2 PM (cutoff 12:30 PM). Previous-day
            // and late-arriving flights are excluded entirely.
            get visibleFlights() {
                const day = this.travel.busto;
                if (!day || day === 'none') return [];
                const busDate = day === 'thursday' ? this.thursdayDate : this.fridayDate;
                if (!busDate) return [];
                const cutoffMinutes = day === 'friday' ? 12 * 60 + 30 : 13 * 60 + 30;
                const q = (this.travel.flightInput || '').toLowerCase().trim();
                return this.allFlights.filter(f => {
                    if (f.localDate !== busDate) return false;
                    const [h, m] = f.localTime.split(':').map(Number);
                    if (h * 60 + m > cutoffMinutes) return false;
                    if (!q) return true;
                    return (f.flight + ' ' + f.airline + ' ' + f.from + ' ' + f.label).toLowerCase().includes(q);
                });
            },

            // Only Sunday departures leaving at or after 13:00 local time — the
            // earliest bus leaves Copán at 8 AM and the drive takes 4-4:15 hours,
            // so earlier flights can't be accommodated on the shared bus.
            get visibleDepartures() {
                if (!this.sundayDate) return [];
                const q = (this.travel.returnDetail || '').toLowerCase().trim();
                return this.allDepartures.filter(f => {
                    if (f.localDate !== this.sundayDate) return false;
                    const [h, m] = f.localTime.split(':').map(Number);
                    if (h < 13) return false;
                    if (!q) return true;
                    return (f.flight + ' ' + f.airline + ' ' + f.to + ' ' + f.label).toLowerCase().includes(q);
                });
            },

            init() {
                const el = this.$el;
                this.lang = el.dataset.lang || 'en';

                const weddingDate = (el.dataset.weddingDate || '').slice(0, 10);
                if (weddingDate) {
                    const date = new Date(weddingDate + 'T12:00:00Z');
                    date.setUTCDate(date.getUTCDate() - 2);
                    this.thursdayDate = date.toISOString().slice(0, 10);
                    date.setUTCDate(date.getUTCDate() + 1);
                    this.fridayDate = date.toISOString().slice(0, 10);
                    date.setUTCDate(date.getUTCDate() + 2);
                    this.sundayDate = date.toISOString().slice(0, 10);
                }

                // Load and normalize flight data (arrivals + Sunday departures)
                try {
                    const flightEl = document.getElementById('sap-flights-data');
                    if (flightEl) {
                        const data = JSON.parse(flightEl.textContent);
                        const SAP_TZ = 'America/Tegucigalpa';
                        const normalize = (f, timestampField) => {
                            const dt = new Date(f[timestampField]);
                            const dateParts = new Intl.DateTimeFormat('en', {
                                timeZone: SAP_TZ, year: 'numeric', month: '2-digit', day: '2-digit'
                            }).formatToParts(dt).reduce((parts, part) => {
                                parts[part.type] = part.value;
                                return parts;
                            }, {});
                            const localDate = `${dateParts.year}-${dateParts.month}-${dateParts.day}`;
                            const localTime = dt.toLocaleTimeString('en-GB', {
                                timeZone: SAP_TZ, hour: '2-digit', minute: '2-digit', hourCycle: 'h23'
                            });
                            const localDateDisplay = dt.toLocaleDateString(this.lang, {
                                timeZone: SAP_TZ, weekday: 'short', month: 'short', day: 'numeric'
                            });
                            // Language-independent format for the value stored/submitted
                            // (shown in the input box and synced to the Google Sheet), e.g. "S085 - 17/12/2026 8:45"
                            const labelTime = dt.toLocaleTimeString('en-GB', {
                                timeZone: SAP_TZ, hour: 'numeric', minute: '2-digit', hourCycle: 'h23'
                            });
                            const label = `${f.flight} - ${dateParts.day}/${dateParts.month}/${dateParts.year} ${labelTime}`;
                            return { ...f, localDate, localTime, localDateDisplay, label };
                        };
                        this.allFlights = (data.arrivals || []).map(f => normalize(f, 'arrives_at'));
                        this.allDepartures = (data.departures || []).map(f => normalize(f, 'departs_at'));
                        // Both lists already sorted by timestamp from YAML
                    }
                } catch(e) { /* non-fatal */ }

                // Load hotel list
                try {
                    const hotelEl = document.getElementById('accommodations-data');
                    if (hotelEl) {
                        this.hotels = JSON.parse(hotelEl.textContent) || [];
                    }
                } catch(e) { /* non-fatal */ }
            },

            setInvite(invite) {
                if (!invite) return;

                this.inHonduras = !!invite.in_honduras;
                const t = this.travel;
                t.busto = invite.travel_bus_to || '';
                t.pickup = invite.travel_pickup || '';
                t.flightInput = invite.travel_arrival_flight || '';
                t.busreturn = invite.travel_bus_return || '';
                t.notes = invite.travel_notes || '';
                t.cocktail = invite.travel_cocktail || '';
                t.brunch = invite.travel_brunch || '';
                t.returnDetail = invite.travel_return_detail || '';

                const hotel = invite.travel_hotel || '';
                if (!hotel || this.hotels.some(h => h.id === hotel)) {
                    t.hotel = hotel;
                    t.hotelOther = '';
                } else {
                    t.hotel = '__other__';
                    t.hotelOther = hotel;
                }

                // Derive two-step return UI state from saved busreturn value
                const br = t.busreturn;
                if (br === 'sunday_morning_sap') {
                    this.returnTime = 'morning';
                    this.returnDest = 'sap';
                } else if (br === 'sunday_morning_san_pedro') {
                    this.returnTime = 'morning';
                    this.returnDest = 'san_pedro';
                } else if (br === 'sunday_afternoon_san_pedro') {
                    this.returnTime = 'afternoon';
                    this.returnDest = 'san_pedro';
                } else if (br === 'none') {
                    this.returnTime = 'none';
                    this.returnDest = '';
                } else {
                    this.returnTime = '';
                    this.returnDest = '';
                }


            },

            // Called when the return bus time changes. The afternoon bus only
            // drops off at San Pedro Sula (no airport stop), so its destination
            // is fixed and doesn't need a separate question.
            onReturnTimeChange() {
                this.returnDest = '';
                this.travel.returnDetail = '';
                if (this.returnTime === 'none') {
                    this.travel.busreturn = 'none';
                } else if (this.returnTime === 'afternoon') {
                    this.returnDest = 'san_pedro';
                    this.travel.busreturn = 'sunday_afternoon_san_pedro';
                } else {
                    // Morning: wait for destination selection before saving a canonical value
                    this.travel.busreturn = '';
                }
                this.scheduleSave();
            },

            // Called when return destination changes (morning bus only)
            onReturnDestChange() {
                this.travel.returnDetail = '';
                if (this.returnTime === 'morning' && this.returnDest) {
                    this.travel.busreturn = 'sunday_morning_' + this.returnDest;
                }
                this.scheduleSave();
            },

            onReturnDetailInput() {
                this.returnFlightOpen = true;
                this.scheduleSave();
            },

            selectReturnFlight(f) {
                this.travel.returnDetail = f.label;
                this.returnFlightOpen = false;
                this.scheduleSave();
            },

            // Called when bus day changes — clear stale dependent state
            onBusToChange() {
                if (this.travel.busto === 'none') {
                    this.travel.pickup = '';
                    this.travel.flightInput = '';
                }
                // If the saved flight is no longer in the visible list, clear it
                if (this.travel.flightInput) {
                    const still = this.visibleFlights.find(f => f.label === this.travel.flightInput);
                    if (!still) this.travel.flightInput = '';
                }
                this.scheduleSave();
            },

            // Called when pickup changes — clear flight if not SAP
            onPickupChange() {
                if (this.travel.pickup !== 'sap') {
                    this.travel.flightInput = '';
                }
                this.scheduleSave();
            },

            onFlightInput() {
                this.flightOpen = true;
                this.scheduleSave();
            },

            selectFlight(f) {
                this.travel.flightInput = f.label;
                this.flightOpen = false;
                this.scheduleSave();
            },

            // Called when hotel radio changes — clear other if switching away
            onHotelChange() {
                if (this.travel.hotel !== '__other__') {
                    this.travel.hotelOther = '';
                }
                this.scheduleSave();
            },

            // Manual save: flush any pending debounce immediately, avoiding duplicate request
            manualSave() {
                if (this.saveTimer) {
                    clearTimeout(this.saveTimer);
                    this.saveTimer = null;
                    this.saveQueued = false;
                }
                this.saveVersion++;
                this.saveStatus = 'saving';
                this.doSave();
            },

            scheduleSave() {
                if (this.saveTimer) {
                    clearTimeout(this.saveTimer);
                }
                this.saveVersion++;
                this.saveStatus = 'saving';
                this.saveTimer = setTimeout(() => {
                    this.saveTimer = null;
                    this.doSave();
                }, 1500);
            },

            async doSave() {
                const code = this.code;
                if (!code) return;
                if (this.saveInFlight) {
                    this.saveQueued = true;
                    return;
                }

                this.saveInFlight = true;
                const version = this.saveVersion;

                const hotelValue = this.travel.hotel === '__other__'
                    ? (this.travel.hotelOther || '')
                    : (this.travel.hotel || '');

                const payload = {
                    bus_to: this.travel.busto,
                    pickup: this.travel.pickup,
                    arrival_flight: this.travel.flightInput,
                    bus_return: this.travel.busreturn,
                    hotel: hotelValue,
                    notes: this.travel.notes,
                    cocktail: this.travel.cocktail,
                    brunch: this.travel.brunch,
                    return_detail: this.travel.returnDetail,
                };

                try {
                    const response = await fetch(`${apiBaseUrl()}/invite/${encodeURIComponent(code)}/travel`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'Accept': 'application/json',
                        },
                        body: JSON.stringify(payload),
                    });
                    if (!response.ok) throw new Error('save failed');
                    if (version === this.saveVersion) {
                        this.saveStatus = 'saved';
                        setTimeout(() => {
                            if (this.saveStatus === 'saved' && version === this.saveVersion) this.saveStatus = '';
                        }, 3000);
                    }
                } catch(e) {
                    if (version === this.saveVersion) {
                        this.saveStatus = 'retry';
                    }
                } finally {
                    this.saveInFlight = false;
                    if (this.saveQueued) {
                        this.saveQueued = false;
                        this.saveStatus = 'saving';
                        this.doSave();
                    }
                }
            },
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

        // -----------------------
        // Schedule Component (fetches and displays wedding schedule)
        // -----------------------
        Alpine.data('scheduleComponent', (lang = 'es') => ({
            loading: true,
            error: null,
            events: [],
            timezone: null,
            timezoneOffset: null,
            lastUpdated: null,
            refreshInterval: null,
            lang: lang, // Language passed from Hugo template
            COPAN_TZ: 'America/Tegucigalpa', // Copan, Honduras timezone (UTC-6, no DST)
            
            init() {
                this.fetchSchedule();
                // Auto-refresh every 30 seconds
                this.refreshInterval = setInterval(() => this.fetchSchedule(), 30000);
            },
            
            destroy() {
                if (this.refreshInterval) {
                    clearInterval(this.refreshInterval);
                }
            },
            
            get apiBase() {
                return apiBaseUrl();
            },
            
            // Get locale string for date formatting
            get locale() {
                const localeMap = { ca: 'ca-ES', es: 'es-ES', en: 'en-US' };
                return localeMap[this.lang] || 'es-ES';
            },
            
            // Get localized text from i18n object { es, en, ca }
            // Falls back to Spanish if the requested language is empty
            localizedText(i18nObj) {
                if (!i18nObj) return '';
                return i18nObj[this.lang] || i18nObj.es || i18nObj.en || i18nObj.ca || '';
            },
            
            async fetchSchedule() {
                const isInitialLoad = this.loading;
                try {
                    const response = await fetch(`${this.apiBase}/schedule`, {
                        headers: { 'Accept': 'application/json' }
                    });
                    
                    if (!response.ok) {
                        throw new Error('Failed to fetch schedule');
                    }
                    
                    const data = await response.json();
                    this.timezone = data.timezone;
                    this.timezoneOffset = data.timezone_offset;
                    this.events = data.events || [];
                    this.lastUpdated = new Date();
                    this.error = null;
                } catch (err) {
                    console.error('Schedule fetch error:', err);
                    // Only set error on initial load; on refresh, keep previous data
                    if (isInitialLoad) {
                        this.error = err.message;
                    }
                } finally {
                    this.loading = false;
                }
            },
            
            // Parse ISO8601 datetime string to Date object
            parseDateTime(isoString) {
                if (!isoString) return null;
                return new Date(isoString);
            },
            
            // Format time in Copan, Honduras timezone (not user's local time)
            // All events happen in Copan so we always show Copan time
            formatLocalTime(isoString) {
                const date = this.parseDateTime(isoString);
                if (!date) return '';
                
                return date.toLocaleTimeString(this.locale, { 
                    hour: '2-digit', 
                    minute: '2-digit',
                    timeZone: this.COPAN_TZ
                });
            },
            
            // Extract date string (YYYY-MM-DD) from ISO datetime for grouping
            // Uses Copan timezone so events group by the correct day
            getDateKey(isoString) {
                const date = this.parseDateTime(isoString);
                if (!date) return '';
                const formatter = new Intl.DateTimeFormat('en-CA', {
                    timeZone: this.COPAN_TZ,
                    year: 'numeric', month: '2-digit', day: '2-digit'
                });
                return formatter.format(date); // en-CA returns YYYY-MM-DD
            },
            
            // Check if an event is in the past
            isEventPast(endTimeIso, startTimeIso) {
                const eventEnd = this.parseDateTime(endTimeIso || startTimeIso);
                if (!eventEnd) return false;
                return eventEnd < new Date();
            },
            
            // Get visible events (filter out past events except the most recent one)
            get visibleEvents() {
                if (!this.events.length) return [];
                
                const processed = this.events.map(event => ({
                    ...event,
                    dateKey: this.getDateKey(event.start_time),
                    localStartTime: this.formatLocalTime(event.start_time),
                    localEndTime: this.formatLocalTime(event.end_time),
                    isPast: this.isEventPast(event.end_time, event.start_time),
                    // Resolve localized name and description
                    localizedName: this.localizedText(event.name),
                    localizedDescription: this.localizedText(event.description)
                }));
                
                // Find the last past event (to show as greyed)
                const pastEvents = processed.filter(e => e.isPast);
                const futureEvents = processed.filter(e => !e.isPast);
                
                // Show the most recent past event (greyed) + all future events
                if (pastEvents.length > 0 && futureEvents.length > 0) {
                    return [pastEvents[pastEvents.length - 1], ...futureEvents];
                }
                
                // If all events are past, show just the last one
                if (pastEvents.length > 0 && futureEvents.length === 0) {
                    return [pastEvents[pastEvents.length - 1]];
                }
                
                return futureEvents;
            },
            
            // Group events by date for display
            get groupedEvents() {
                const groups = {};
                for (const event of this.visibleEvents) {
                    if (!groups[event.dateKey]) {
                        groups[event.dateKey] = [];
                    }
                    groups[event.dateKey].push(event);
                }
                return groups;
            },
            
            // Format date header (e.g., "Friday, December 19")
            // dateStr is YYYY-MM-DD from getDateKey (already in Copan TZ)
            formatDateHeader(dateStr) {
                const date = new Date(dateStr + 'T12:00:00');
                return date.toLocaleDateString(this.locale, {
                    weekday: 'long',
                    month: 'long',
                    day: 'numeric'
                });
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
