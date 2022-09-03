package config

const FLASH_QUERY_ADDRESS = "0x9C7FfE06A4c5C58A5D60bC95baAb56F558A4dACf" // Mumbai
const BUNDLE_EXECUTOR_ADDRESS = "0x8Bdc9d868950E6993B2d0Aa8e56E1cEDa5140200"

const WETH_ADDRESS = "0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270" // Mumbai
// const WETH_ADDRESS = "0x7ceb23fd6bc0add59e62ac25578270cff1b9f619" // Polygon

const QUICKSWAP_FACTORY_ADDRESS = "0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32" // Mumbai and Polygon
const SUSHISWAP_FACTORY_ADDRESS = "0xc35DADB65012eC5796536bD9864eD8773aBc74C4" // Mumbai and Polygon

var UNISWAPPY_FACTORY_ADDRESSES = []string{SUSHISWAP_FACTORY_ADDRESS, QUICKSWAP_FACTORY_ADDRESS}