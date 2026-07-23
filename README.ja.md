# jpeg2000

[![Go Reference](https://pkg.go.dev/badge/github.com/d-fuji/jpeg2000.svg)](https://pkg.go.dev/github.com/d-fuji/jpeg2000)

[English](README.md) | 日本語

JPEG 2000(ISO/IEC 15444-1)画像の純 Go 実装デコーダです。cgo も外部
ライブラリも使用していません — MQ 算術符号器から逆ウェーブレット変換
まで、コーデック全体を Go で実装しています。

```
go get github.com/d-fuji/jpeg2000
```

## 使い方

標準ライブラリの `image` パッケージにそのまま組み込めます:

```go
import (
    "image"
    _ "github.com/d-fuji/jpeg2000"
)

f, _ := os.Open("photo.jp2")
img, _, err := image.Decode(f)
```

直接呼び出すこともできます:

```go
img, err := jpeg2000.Decode(f)
cfg, err := jpeg2000.DecodeConfig(f) // 寸法とカラーモデルのみ取得
```

JP2 コンテナ形式(`.jp2`)と生のコードストリーム
(`.j2k`/`.j2c`/`.jpc`)の両方に対応しています。

## 機能

- 可逆(5/3)・非可逆(9/7)ウェーブレット変換 —
  ロスレス画像はビット完全一致でデコード
- マルチタイル、タイルパート、品質レイヤ、解像度レベル
- 全 5 種の進行順序(LRCP、RLCP、RPCL、PCRL、CPRL)
- 任意のプリシンクト / コードブロックサイズ、SOP/EPH マーカー
- RCT/ICT 複数コンポーネント(色)変換
- グレースケール、RGB、アルファチャンネル画像(コンポーネントあたり
  最大 16 ビット)
- コードブロックスタイル: 垂直因果コンテキスト、予測可能終端、
  パスごとの終端、コンテキストリセット、セグメンテーションシンボル

未対応の機能([ROADMAP.ja.md](ROADMAP.ja.md) 参照): 算術符号器
バイパスモード、POC(進行順序変更)、ROI(RGN)、パレット画像、
コンポーネントのサブサンプリング、HTJ2K(Part 15)。

## 正当性の検証

テストスイートは、機能マトリクス全域(変換、進行順序、タイル分割、
プリシンクト、レイヤ)にわたって OpenJPEG でエンコードした画像を
デコードし、OpenJPEG 自身のデコード結果と比較します: ロスレス画像は
ビット完全一致、ロッシー画像は ±1 階調以内(浮動小数点丸め誤差)です。
不正な入力に対するファジングも行っています。

テスト画像の再生成には Python と Pillow が必要です:

```
python3 testdata/gen.py
```

## ステータス

このライブラリはまだ若いプロジェクトです。テストマトリクス上のすべてで
正しく動作しますが、速度の最適化はこれからで、ISO 適合性スイート全体に
対する検証もまだです。サンプルファイル付きのバグ報告を歓迎します。

## ライセンス

Apache License 2.0。[LICENSE](LICENSE) を参照してください。
