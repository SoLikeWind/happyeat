import { Component, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

/** 捕获子组件抛出的错误，避免整页白屏，并显示错误信息 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  render() {
    if (this.state.hasError && this.state.error) {
      return (
        <div style={{ padding: 24, fontFamily: 'sans-serif' }}>
          <h2 style={{ color: '#c00' }}>页面出错了</h2>
          <pre style={{ background: '#f5f5f5', padding: 12, overflow: 'auto' }}>
            {this.state.error.message}
          </pre>
          <p style={{ color: '#666' }}>请打开浏览器开发者工具 (F12) 查看 Console 获取完整报错。</p>
        </div>
      )
    }
    return this.props.children
  }
}
