module.exports = {
    mode: 'production',
    module: {
        rules: [
            {
                test: /\.scss$/i,
                use: [
                    { loader: MiniCssExtractPlugin.loader, options: {} },
                    {
                        loader: 'css-loader',
                        options: { url: false },
                    },
                ],
            },
        ],
    },
    optimization: {
        minimizer: [new CssMinimizerPlugin(), new TerserPlugin()],
    },
    plugins: [
        new MiniCssExtractPlugin({ filename: `app.css` }),
        new CleanWebpackPlugin(),
        new HtmlWebpackPlugin({
            inject: true,
            filename: 'index.html',
            template: 'src/static/index.template.html',
            minify: false,
        }),
    ],
}
