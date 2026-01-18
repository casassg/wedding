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
        const browserLangs = navigator.languages || [navigator.language || 'en'];
        let preferredLang = 'en'; // default fallback

        for (const lang of browserLangs) {
            const code = lang.toLowerCase().split('-')[0];
            if (code === 'es') { preferredLang = 'es'; break; }
            if (code === 'ca') { preferredLang = 'ca'; break; }
            if (code === 'en') { preferredLang = 'en'; break; }
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
    // Initialize Everything
    // ===================
    document.addEventListener('DOMContentLoaded', () => {
        initMobileMenu();
        initScrollAnimations();
        initSmoothScroll();
        initNavScroll();
    });

})();
