import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../core/auth.service';
import { MatDialog } from '@angular/material/dialog';
import { TwoFADialogComponent } from '../core/twofa-dialog.component';

@Component({
  selector: 'app-login',
  template: `
    <div style="min-height: 100vh; background: linear-gradient(135deg, var(--primary-dark), var(--primary)); display: flex; align-items: center; justify-content: center; padding: 20px;">
      <div style="background: white; border-radius: 12px; box-shadow: 0 20px 60px rgba(0,0,0,0.3); padding: 40px; max-width: 420px; width: 100%;">
        <div style="text-align: center; margin-bottom: 32px;">
          <div style="width: 60px; height: 60px; background: linear-gradient(135deg, var(--primary), var(--primary-light)); border-radius: 12px; display: inline-flex; align-items: center; justify-content: center; font-size: 32px; margin-bottom: 16px;">🏦</div>
          <h1 class="bank-title" style="margin-bottom: 8px;">银联银行</h1>
          <p style="color: #666;">个人网上银行</p>
        </div>

        <div *ngIf="error" class="alert alert-error">{{ error }}</div>

        <form (ngSubmit)="onSubmit()">
          <div class="bank-form-group">
            <label class="bank-label">用户名</label>
            <input type="text" class="bank-input" [(ngModel)]="username" name="username" required placeholder="请输入用户名" />
          </div>

          <div class="bank-form-group">
            <label class="bank-label">密码</label>
            <input type="password" class="bank-input" [(ngModel)]="password" name="password" required placeholder="请输入密码" />
          </div>

          <button type="submit" class="bank-btn bank-btn-primary" style="width: 100%; margin-bottom: 16px;" [disabled]="loading">
            {{ loading ? '登录中...' : '登 录' }}
          </button>
        </form>

        <div style="text-align: center; color: #666;">
          还没有账户？<a routerLink="/register" style="color: var(--primary); text-decoration: none;">立即注册</a>
        </div>
      </div>
    </div>
  `
})
export class LoginComponent {
  username = '';
  password = '';
  error = '';
  loading = false;

  constructor(
    private authService: AuthService,
    private router: Router,
    private dialog: MatDialog
  ) {}

  onSubmit(): void {
    if (!this.username || !this.password) {
      this.error = '请输入用户名和密码';
      return;
    }

    this.loading = true;
    this.error = '';

    this.authService.login(this.username, this.password).subscribe({
      next: (resp) => {
        this.loading = false;
        if (resp.need_twofa && resp.twofa_token) {
          this.openTwoFADialog(resp.twofa_token);
        } else if (resp.token && resp.user) {
          this.authService.setAuth(resp.token, resp.user);
          this.router.navigate(['/accounts']);
        }
      },
      error: (err) => {
        this.loading = false;
        this.error = err.error?.error || '登录失败，请重试';
      }
    });
  }

  openTwoFADialog(twoFAToken: string): void {
    const dialogRef = this.dialog.open(TwoFADialogComponent, {
      width: '400px',
      data: { target: '您的手机', action: 'login' }
    });

    dialogRef.afterClosed().subscribe((code: string | null) => {
      if (code) {
        this.verifyTwoFA(twoFAToken, code);
      }
    });
  }

  verifyTwoFA(token: string, code: string): void {
    this.authService.verifyTwoFA(token, code).subscribe({
      next: (resp) => {
        if (resp.token && resp.user) {
          this.authService.setAuth(resp.token, resp.user);
          this.router.navigate(['/accounts']);
        }
      },
      error: (err) => {
        this.error = err.error?.error || '验证失败';
      }
    });
  }
}
