import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../core/auth.service';

@Component({
  selector: 'app-register',
  template: `
    <div style="min-height: 100vh; background: linear-gradient(135deg, var(--primary-dark), var(--primary)); display: flex; align-items: center; justify-content: center; padding: 20px;">
      <div style="background: white; border-radius: 12px; box-shadow: 0 20px 60px rgba(0,0,0,0.3); padding: 40px; max-width: 500px; width: 100%;">
        <h2 class="bank-title" style="text-align: center; margin-bottom: 24px;">注册新账户</h2>

        <div *ngIf="error" class="alert alert-error">{{ error }}</div>
        <div *ngIf="success" class="alert alert-success">注册成功！正在跳转到登录页...</div>

        <form (ngSubmit)="onSubmit()">
          <div class="grid-2">
            <div class="bank-form-group">
              <label class="bank-label">用户名</label>
              <input type="text" class="bank-input" [(ngModel)]="form.username" name="username" required />
            </div>
            <div class="bank-form-group">
              <label class="bank-label">姓名</label>
              <input type="text" class="bank-input" [(ngModel)]="form.full_name" name="full_name" required />
            </div>
          </div>

          <div class="bank-form-group">
            <label class="bank-label">密码</label>
            <input type="password" class="bank-input" [(ngModel)]="form.password" name="password" required minlength="8" />
            <small style="color: #999;">至少8位字符</small>
          </div>

          <div class="grid-2">
            <div class="bank-form-group">
              <label class="bank-label">手机号</label>
              <input type="tel" class="bank-input" [(ngModel)]="form.phone" name="phone" required />
            </div>
            <div class="bank-form-group">
              <label class="bank-label">邮箱</label>
              <input type="email" class="bank-input" [(ngModel)]="form.email" name="email" required />
            </div>
          </div>

          <div class="bank-form-group">
            <label class="bank-label">身份证号</label>
            <input type="text" class="bank-input" [(ngModel)]="form.id_card" name="id_card" required />
          </div>

          <button type="submit" class="bank-btn bank-btn-gold" style="width: 100%; margin-bottom: 16px;" [disabled]="loading">
            {{ loading ? '注册中...' : '立即注册' }}
          </button>
        </form>

        <div style="text-align: center; color: #666;">
          已有账户？<a routerLink="/login" style="color: var(--primary); text-decoration: none;">返回登录</a>
        </div>
      </div>
    </div>
  `
})
export class RegisterComponent {
  form = {
    username: '',
    password: '',
    full_name: '',
    phone: '',
    email: '',
    id_card: ''
  };
  error = '';
  success = false;
  loading = false;

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  onSubmit(): void {
    this.loading = true;
    this.error = '';

    this.authService.register(this.form).subscribe({
      next: () => {
        this.success = true;
        setTimeout(() => {
          this.router.navigate(['/login']);
        }, 2000);
      },
      error: (err) => {
        this.loading = false;
        this.error = err.error?.error || '注册失败，请重试';
      }
    });
  }
}
