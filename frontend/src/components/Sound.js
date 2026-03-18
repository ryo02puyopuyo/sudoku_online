import correctWav from '../sounds/決定ボタンを押す52.mp3';
import incorrectWav from '../sounds/ビープ音4.mp3';
import hotspotWav from '../sounds/決定ボタンを押す37.mp3';

// ========================================================
// ハイブリッド音声再生モジュール
// ========================================================
// 初回クリック: HTML5 Audio で再生（確実にユーザージェスチャー内）
// 2回目以降 : Web Audio API で低遅延・重複再生
//
// ※ Firefoxでは初回クリック時にブラウザ自体の遅延が発生するが、
//    これはブラウザの仕様であり回避不可。
// ========================================================

const soundUrls = {
  correct: correctWav,
  incorrect: incorrectWav,
  hotspot: hotspotWav,
};

const volumes = {
  correct: 1.0,
  incorrect: 0.5,
  hotspot: 0.8,
};

// --- Web Audio API 関連 ---
let audioCtx = null;
const audioBuffers = {};

// AudioContextの初期化（同期。awaitしない）
const initAudioContext = () => {
  if (audioCtx) return;
  audioCtx = new (window.AudioContext || window.webkitAudioContext)();

  // 事前にfetchした生バイナリをバックグラウンドでデコード
  Object.entries(rawData).forEach(([name, buf]) => {
    if (buf) {
      audioCtx.decodeAudioData(buf)
        .then((decoded) => { audioBuffers[name] = decoded; })
        .catch((e) => console.error(`デコード失敗: ${name}`, e));
    }
  });
};

// --- 生バイナリの事前取得 ---
const rawData = {};
Object.entries(soundUrls).forEach(([name, url]) => {
  fetch(url)
    .then((res) => res.arrayBuffer())
    .then((buf) => { rawData[name] = buf; })
    .catch((e) => console.error(`取得失敗: ${name}`, e));
});

// --- 再生 ---
export const playSE = (key) => {
  // AudioContextを初期化（初回のみ、同期的）
  initAudioContext();

  // Web Audio APIで再生（デコード済みバッファがある場合）
  const buffer = audioBuffers[key];
  if (buffer && audioCtx && audioCtx.state === 'running') {
    const source = audioCtx.createBufferSource();
    source.buffer = buffer;
    const gainNode = audioCtx.createGain();
    gainNode.gain.value = volumes[key] || 1.0;
    source.connect(gainNode);
    gainNode.connect(audioCtx.destination);
    source.start(audioCtx.currentTime);
    return;
  }

  // フォールバック: HTML5 Audio（初回クリック時）
  const url = soundUrls[key];
  if (url) {
    const audio = new Audio(url);
    audio.volume = volumes[key] || 1.0;
    audio.play().catch((e) => console.log('再生エラー:', e));
    audio.addEventListener('ended', () => { audio.src = ''; });
  }
};