/**
 * Laura & Gerard Wedding - Main JavaScript
 * Lightweight, vanilla JS for countdown, FAQ, and animations
 */

(function() {
    'use strict';

    // ===================
    // Countdown Timer
    // ===================
    const weddingDate = new Date("Dec 19, 2026 16:00:00").getTime();

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
    // FAQ Accordion
    // ===================
    function initFAQ() {
        const faqItems = document.querySelectorAll('.faq-item');
        
        // Remove hidden class on init - CSS handles visibility via max-height
        faqItems.forEach(item => {
            const answer = item.querySelector('.faq-answer');
            if (answer) {
                answer.classList.remove('hidden');
            }
        });
        
        faqItems.forEach(item => {
            const question = item.querySelector('.faq-question');
            
            if (question) {
                question.addEventListener('click', () => {
                    // Close all other items
                    faqItems.forEach(otherItem => {
                        if (otherItem !== item) {
                            otherItem.classList.remove('active');
                        }
                    });
                    
                    // Toggle current item
                    item.classList.toggle('active');
                });
            }
        });
    }

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
        initFAQ();
        initMobileMenu();
        initScrollAnimations();
        initSmoothScroll();
        initNavScroll();
    });

})();
