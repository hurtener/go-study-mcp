<script>
  let { onGenerating } = $props();

  let text = $state('');
  let voice = $state('alloy');
  let format = $state('mp3');

  const voices = [
    { id: 'alloy', label: 'Alloy', desc: 'Neutral, balanced' },
    { id: 'echo', label: 'Echo', desc: 'Clear, articulate' },
    { id: 'fable', label: 'Fable', desc: 'Warm, expressive' },
    { id: 'onyx', label: 'Onyx', desc: 'Deep, authoritative' },
    { id: 'nova', label: 'Nova', desc: 'Friendly, upbeat' },
    { id: 'shimmer', label: 'Shimmer', desc: 'Soft, gentle' },
  ];

  async function handleSubmit(e) {
    e.preventDefault();
    if (!text.trim()) return;
    onGenerating();
  }
</script>

<form class="synthesize-form" onsubmit={handleSubmit}>
  <div class="field">
    <label for="syn-text">Text to Synthesize</label>
    <textarea
      id="syn-text"
      bind:value={text}
      placeholder="Type or paste text here. Use [PAUSE:5] for a 5-second pause..."
      rows="6"
      required
    ></textarea>
    <span class="hint">
      Tip: Use [PAUSE:N] markers to insert N-second silences
    </span>
  </div>

  <div class="row">
    <div class="field">
      <label for="syn-voice">Voice</label>
      <div class="voice-grid">
        {#each voices as v}
          <button
            type="button"
            class="voice-btn"
            class:active={voice === v.id}
            onclick={() => voice = v.id}
          >
            <span class="voice-name">{v.label}</span>
            <span class="voice-desc">{v.desc}</span>
          </button>
        {/each}
      </div>
    </div>

    <div class="field">
      <label for="syn-format">Format</label>
      <div class="format-toggle">
        <button
          type="button"
          class="format-btn"
          class:active={format === 'mp3'}
          onclick={() => format = 'mp3'}
        >
          MP3
        </button>
        <button
          type="button"
          class="format-btn"
          class:active={format === 'wav'}
          onclick={() => format = 'wav'}
        >
          WAV
        </button>
      </div>
    </div>
  </div>

  <div class="actions">
    <div class="char-count">
      {text.length > 0 ? `${text.replace(/\[PAUSE:\d+\]/g, '').length} chars` : ''}
    </div>
    <button type="submit" class="btn-primary" disabled={!text.trim()}>
      Synthesize Speech
    </button>
  </div>
</form>

<style>
  .synthesize-form {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  label {
    font-size: 12px;
    font-weight: 600;
    color: #4a4a4a;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  textarea {
    padding: 14px;
    border: 1px solid #e8e5e0;
    border-radius: 12px;
    font-size: 15px;
    font-family: inherit;
    background: #fff;
    color: #2d2d2d;
    resize: vertical;
    min-height: 140px;
    line-height: 1.6;
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  textarea:focus {
    outline: none;
    border-color: #8b7355;
    box-shadow: 0 0 0 3px rgba(139, 115, 85, 0.1);
  }

  textarea::placeholder {
    color: #b0a89e;
  }

  .hint {
    font-size: 11px;
    color: #999;
    font-style: italic;
  }

  .row {
    display: grid;
    grid-template-columns: 1fr;
    gap: 16px;
  }

  @media (min-width: 480px) {
    .row {
      grid-template-columns: 2fr 1fr;
    }
  }

  .voice-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 6px;
  }

  .voice-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 10px 6px;
    border: 1px solid #e8e5e0;
    border-radius: 10px;
    background: #fff;
    cursor: pointer;
    transition: all 0.2s;
  }

  .voice-btn:hover {
    border-color: #d0cbc3;
  }

  .voice-btn.active {
    border-color: #8b7355;
    background: #f9f5ee;
  }

  .voice-name {
    font-size: 13px;
    font-weight: 600;
    color: #2d2d2d;
  }

  .voice-desc {
    font-size: 10px;
    color: #8a8a8a;
  }

  .format-toggle {
    display: flex;
    gap: 4px;
    background: #f5f3f0;
    padding: 3px;
    border-radius: 8px;
  }

  .format-btn {
    flex: 1;
    padding: 10px;
    border: none;
    background: transparent;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
    color: #6b6b6b;
    cursor: pointer;
    transition: all 0.2s;
  }

  .format-btn.active {
    background: #fff;
    color: #1a1a1a;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
  }

  .actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding-top: 4px;
  }

  .char-count {
    font-size: 12px;
    color: #999;
  }

  .btn-primary {
    padding: 12px 24px;
    background: linear-gradient(135deg, #8b7355 0%, #a08060 100%);
    color: #fff;
    border: none;
    border-radius: 10px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 2px 8px rgba(139, 115, 85, 0.25);
  }

  .btn-primary:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(139, 115, 85, 0.35);
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
