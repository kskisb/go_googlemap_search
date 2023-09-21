# go_googlemap_search

Google Map上の飲食店を人気順(※)に表示するLINE Botです。

## 使い方
下の写真の手順通りとなっています。

<img width="1223" alt="image" src="https://github.com/kskisb/go_googlemap_search/assets/134213401/5d061446-2a93-48cc-96b7-92b9dc6fff88">


## 人気順(※)について
このBotではGoogle Mapにおける人気店の定義を

平均の星評価^(2.5)*ユーザーからの星評価の総数^(0.25)

という計算式で定義しました。この式は、口コミ件数がそれほど多くない場合でも平均の星評価が高い場合にはランキングに載りやすくなることを目的として設定しました。  
あくまでも私が設定した式の上での人気順ですので、参考程度に受け止めていただけたらと思います。

## 参考
[GoとLINE BOTにまとめて入門する](https://speakerdeck.com/yagieng/go-to-line-bot-ni-matometeru-men-suru)
