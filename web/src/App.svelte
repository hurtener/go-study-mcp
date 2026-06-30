<script>
  import { onMount } from 'svelte';
  import { createBridge } from 'dockyard-bridge';
  import PodcastView from './lib/components/PodcastView.svelte';
  import FlashcardView from './lib/components/FlashcardView.svelte';
  import SynthesizeView from './lib/components/SynthesizeView.svelte';
  import StudyGuideView from './lib/components/StudyGuideView.svelte';
  import JobsView from './lib/components/JobsView.svelte';
  import AudioPlayer from './lib/components/AudioPlayer.svelte';

  let activeTab = $state('podcast');
  let toolResult = $state(null);
  let isGenerating = $state(false);
  let ready = $state(false);
  let error = $state(null);
  let permission = $state(false);

  const bridge = createBridge({ displayModes: ['inline'] });

  const tabs = [
    { id: 'podcast', label: 'Podcast', icon: '🎙️' },
    { id: 'study_guide', label: 'Study Guide', icon: '📖' },
    { id: 'flashcards', label: 'Flashcards', icon: '🃏' },
    { id: 'synthesize', label: 'Synthesize', icon: '🔊' },
    { id: 'jobs', label: 'Jobs', icon: '📂' },
  ];

  bridge.onToolResult((payload) => {
    const sc = payload.structuredContent;
    isGenerating = false;
    // Async audio generation returns a job handle ("processing"); the actual
    // audio shows up in the Jobs tab. Preview results render inline below.
    if (sc && sc.status === 'processing') {
      toolResult = null;
      activeTab = 'jobs';
      return;
    }
    toolResult = sc;
  });

  onMount(async () => {
    try {
      await bridge.connect();
      ready = true;
    } catch (e) {
      error = e.message || 'Failed to connect to host';
    }
  });

  function handleGenerating() {
    isGenerating = true;
    toolResult = null;
  }
</script>

<div class="app">
  {#if !ready && !error && !permission}
    <!-- Loading State -->
    <div class="state loading">
      <div class="state-icon">
        <div class="spinner"></div>
      </div>
      <h2>Connecting...</h2>
      <p>Setting up Study Audio Studio</p>
    </div>
  {:else if error}
    <!-- Error State -->
    <div class="state error">
      <div class="state-icon">⚠️</div>
      <h2>Connection Error</h2>
      <p>{error}</p>
      <button class="btn-retry" onclick={() => window.location.reload()}>
        Retry
      </button>
    </div>
  {:else if permission}
    <!-- Permission State -->
    <div class="state permission">
      <div class="state-icon">🔐</div>
      <h2>Configuration Required</h2>
      <p>This tool requires an API key to generate audio. Please configure your OPENAI_API_KEY environment variable.</p>
      <code class="config-hint">OPENAI_API_KEY=sk-your-key-here</code>
    </div>
  {:else if !toolResult && !isGenerating}
    <!-- Empty State -->
    <div class="empty">
      <header class="header">
        <div class="logo">
          <span class="logo-icon">📚</span>
          <h1>Study Audio Studio</h1>
        </div>
        <p class="subtitle">Transform your study material into audio</p>
      </header>

      <nav class="tabs">
        {#each tabs as tab}
          <button
            class="tab"
            class:active={activeTab === tab.id}
            onclick={() => activeTab = tab.id}
          >
            <span class="tab-icon">{tab.icon}</span>
            <span class="tab-label">{tab.label}</span>
          </button>
        {/each}
      </nav>

      <main class="content">
        {#if activeTab === 'podcast'}
          <PodcastView onGenerating={handleGenerating} />
        {:else if activeTab === 'study_guide'}
          <StudyGuideView onGenerating={handleGenerating} />
        {:else if activeTab === 'flashcards'}
          <FlashcardView onGenerating={handleGenerating} />
        {:else if activeTab === 'synthesize'}
          <SynthesizeView onGenerating={handleGenerating} />
        {:else if activeTab === 'jobs'}
          <JobsView {bridge} />
        {/if}
      </main>
    </div>
  {:else}
    <!-- Ready State with result -->
    <div class="ready">
      <header class="header compact">
        <div class="logo">
          <span class="logo-icon">📚</span>
          <h1>Study Audio Studio</h1>
        </div>
      </header>

      {#if isGenerating}
        <div class="generating">
          <div class="spinner"></div>
          <span>Generating audio...</span>
        </div>
      {:else if toolResult}
        <AudioPlayer result={toolResult} />
        <button class="btn-new" onclick={() => { toolResult = null; }}>
          Create Another
        </button>
      {/if}
    </div>
  {/if}
</div>

<style>
  :global(*) {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }

  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
    background: linear-gradient(135deg, #fefefe 0%, #faf8f5 100%);
    color: #2d2d2d;
    line-height: 1.5;
  }

  .app {
    max-width: 640px;
    margin: 0 auto;
    padding: 24px 20px;
    min-height: 100vh;
  }

  .state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 60px 20px;
    gap: 12px;
  }

  .state-icon {
    font-size: 48px;
    margin-bottom: 8px;
  }

  .state h2 {
    font-size: 18px;
    font-weight: 600;
    color: #1a1a1a;
  }

  .state p {
    font-size: 14px;
    color: #6b6b6b;
  }

  .btn-retry {
    margin-top: 12px;
    padding: 10px 20px;
    background: #8b7355;
    color: #fff;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
  }

  .config-hint {
    margin-top: 12px;
    padding: 8px 16px;
    background: #f5f3f0;
    border-radius: 6px;
    font-size: 12px;
    font-family: 'SF Mono', Monaco, monospace;
    color: #6b6b6b;
  }

  .empty, .ready {
    display: flex;
    flex-direction: column;
  }

  .header {
    text-align: center;
    margin-bottom: 28px;
  }

  .header.compact {
    margin-bottom: 16px;
  }

  .logo {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    margin-bottom: 6px;
  }

  .logo-icon {
    font-size: 28px;
  }

  h1 {
    font-size: 22px;
    font-weight: 600;
    color: #1a1a1a;
    letter-spacing: -0.3px;
  }

  .subtitle {
    font-size: 13px;
    color: #7a7a7a;
  }

  .tabs {
    display: flex;
    gap: 6px;
    background: #f5f3f0;
    padding: 4px;
    border-radius: 12px;
    margin-bottom: 24px;
  }

  .tab {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 10px 12px;
    border: none;
    background: transparent;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s ease;
    font-size: 13px;
    font-weight: 500;
    color: #6b6b6b;
  }

  .tab:hover {
    background: rgba(255, 255, 255, 0.6);
    color: #3d3d3d;
  }

  .tab.active {
    background: #ffffff;
    color: #1a1a1a;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  }

  .tab-icon {
    font-size: 15px;
  }

  .tab-label {
    display: none;
  }

  @media (min-width: 480px) {
    .tab-label {
      display: inline;
    }
  }

  .content {
    position: relative;
  }

  .generating {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    padding: 20px;
    background: #f8f6f3;
    border-radius: 12px;
    color: #6b6b6b;
    font-size: 13px;
  }

  .spinner {
    width: 16px;
    height: 16px;
    border: 2px solid #e0ddd8;
    border-top-color: #8b7355;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  .btn-new {
    margin-top: 16px;
    padding: 12px 24px;
    background: transparent;
    border: 1px solid #e8e5e0;
    border-radius: 10px;
    font-size: 14px;
    font-weight: 500;
    color: #6b6b6b;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-new:hover {
    background: #f5f3f0;
    color: #3d3d3d;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
