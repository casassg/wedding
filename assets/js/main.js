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

        // Only run on homepage URLs
        const path = window.location.pathname;
        const homepagePaths = ['/wedding/', '/wedding/en/', '/wedding/es/', '/wedding/ca/'];
        const isHomepage = homepagePaths.includes(path) || path === '/wedding';
        if (!isHomepage) return;

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
            const baseUrl = '/wedding/';
            const newUrl = preferredLang === 'en' ? baseUrl : baseUrl + preferredLang + '/';
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
    // Place Cards Modal
    // ===================
    function initPlaceCards() {
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
        document.querySelectorAll('.place-modal-backdrop').forEach(backdrop => {
            backdrop.addEventListener('click', () => {
                const modal = backdrop.closest('.place-modal');
                if (modal) {
                    modal.classList.remove('active');
                    document.body.classList.remove('place-modal-open');
                }
            });
        });

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

        // Close modal on Escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                const activeModal = document.querySelector('.place-modal.active');
                if (activeModal) {
                    activeModal.classList.remove('active');
                    document.body.classList.remove('place-modal-open');
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
    });

})();
