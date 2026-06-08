import { Component, OnInit } from '@angular/core';
import { Router, NavigationEnd } from '@angular/router';
import { AuthService } from './core/auth.service';
import { filter } from 'rxjs';

@Component({
  selector: 'app-root',
  template: `
    <div *ngIf="authService.isLoggedIn()">
      <header class="bank-header">
        <div class="bank-header-content">
          <div class="bank-logo">
            <div class="bank-logo-icon">🏦</div>
            <span>银联银行</span>
          </div>
          <nav class="bank-nav">
            <a routerLink="/accounts" routerLinkActive="active">我的账户</a>
            <a routerLink="/transfer" routerLinkActive="active">转账汇款</a>
            <a routerLink="/transactions" routerLinkActive="active">交易明细</a>
            <a routerLink="/recon" routerLinkActive="active">对账中心</a>
            <a routerLink="/audit" routerLinkActive="active">审计日志</a>
            <a routerLink="/settings" routerLinkActive="active">设置</a>
          </nav>
          <div style="display: flex; align-items: center; gap: 16px;">
            <span style="color: var(--text-secondary);">
              欢迎，{{ currentUser?.full_name || '用户' }}
            </span>
            <button class="bank-btn bank-btn-gold" style="padding: 8px 16px; font-size: 14px;" (click)="logout()">
              退出
            </button>
          </div>
        </div>
      </header>
      <main class="bank-container" style="padding-top: 24px;">
        <router-outlet></router-outlet>
      </main>
    </div>
    <div *ngIf="!authService.isLoggedIn()">
      <router-outlet></router-outlet>
    </div>
  `
})
export class AppComponent implements OnInit {
  currentUser: any = null;

  constructor(
    public authService: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.authService.user$.subscribe(user => {
      this.currentUser = user;
    });

    this.router.events.pipe(
      filter(event => event instanceof NavigationEnd)
    ).subscribe(() => {
      if (!this.authService.isLoggedIn() && this.router.url !== '/login' && this.router.url !== '/register') {
        this.router.navigate(['/login']);
      }
    });
  }

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}
