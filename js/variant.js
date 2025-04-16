window.relearn = window.relearn || {};

// we need to load this script in the html head to avoid flickering
// on page load if the user has selected a non default variant

function ready(fn) { if (document.readyState == 'complete') { fn(); } else { document.addEventListener('DOMContentLoaded',fn); } }

var variants = {
	variant: '',
	variants: [],
	customvariantname: 'my-custom-variant',
	isstylesheetloaded: true,

	init: function( variants ){
		this.variants = variants;
		var variant = window.localStorage.getItem( window.relearn.absBaseUri+'/variant' ) || ( this.variants.length ? this.variants[0] : '' );
		this.changeVariant( variant );
		document.addEventListener( 'readystatechange', function(){
			if( document.readyState == 'interactive' ){
				this.markSelectedVariant();
			}
		}.bind( this ) );
	},

	getVariant: function(){
		return this.variant;
	},

	setVariant: function( variant ){
		this.variant = variant;
		window.localStorage.setItem( window.relearn.absBaseUri+'/variant', variant );
	},

	isVariantLoaded: function(){
		return window.theme && this.isstylesheetloaded;
	},

	markSelectedVariant: function(){
		var variant = this.getVariant();
		var select = document.querySelector( '#R-select-variant' );
		if( !select ){
			return;
		}
		this.addCustomVariantOption();
		if( variant && select.value != variant ){
			select.value = variant;
		}
		var interval_id = setInterval( function(){
			if( this.isVariantLoaded() ){
				clearInterval( interval_id );
				updateTheme({ variant: variant });
			}
		}.bind( this ), 25 );
		// remove selection, because if some uses an arrow navigation"
		// by pressing the left or right cursor key, we will automatically
		// select a different style
		if( document.activeElement ){
			document.activeElement.blur();
		}
	},

	generateVariantPath: function( variant, old_path ){
		var new_path = old_path.replace( new RegExp(`^(.*\/theme-).*?(\.css.*)$`), '$1' + variant + '$2' );
		return new_path;
	},

	addCustomVariantOption: function(){
		var variantbase = window.localStorage.getItem( window.relearn.absBaseUri+'/customvariantbase' );
		if( this.variants.indexOf( variantbase ) < 0 ){
			variantbase = '';
		}
		if( !window.localStorage.getItem( window.relearn.absBaseUri+'/customvariant' ) ){
			variantbase = '';
		}
		if( !variantbase ){
			return;
		}
		var select = document.querySelector( '#R-select-variant' );
		if( !select ){
			return;
		}
		var option = document.querySelector( '#' + this.customvariantname );
		if( !option ){
			option = document.createElement( 'option' );
			option.id = this.customvariantname;
			option.value = this.customvariantname;
			option.text = this.customvariantname.replace( /-/g, ' ' ).replace(/\w\S*/g, function(w){ return w.replace(/^\w/g, function(c){ return c.toUpperCase(); }); });
			select.appendChild( option );
			document.querySelectorAll( '.footerVariantSwitch' ).forEach( function( e ){
				e.classList.add( 'showVariantSwitch' );
			});
		}
	},

	removeCustomVariantOption: function(){
		var option = document.querySelector( '#' + this.customvariantname );
		if( option ){
			option.remove();
		}
		if( this.variants.length <= 1 ){
			document.querySelectorAll( '.footerVariantSwitch' ).forEach( function( e ){
				e.classList.remove( 'showVariantSwitch' );
			});
		}
	},

	saveCustomVariant: function(){
		if( this.getVariant() != this.customvariantname ){
			window.localStorage.setItem( window.relearn.absBaseUri+'/customvariantbase', this.getVariant() );
		}
		window.localStorage.setItem( window.relearn.absBaseUri+'/customvariant', this.generateStylesheet() );
		this.setVariant( this.customvariantname );
		this.markSelectedVariant();
	},

	loadCustomVariant: function(){
		var stylesheet = window.localStorage.getItem( window.relearn.absBaseUri+'/customvariant' );

		// temp styles to document
		var head = document.querySelector( 'head' );
		var style = document.createElement( 'style' );
		style.id = 'R-custom-variant-style';
		style.appendChild( document.createTextNode( stylesheet ) );
		head.appendChild( style );

		var interval_id = setInterval( function(){
			if( this.findLoadedStylesheet( 'R-variant-style' ) ){
				clearInterval( interval_id );
				// save the styles to the current variant stylesheet
				this.variantvariables.forEach( function( e ){
					this.changeColor( e.name, true );
				}.bind( this ) );

				// remove temp styles
				style.remove();

				this.saveCustomVariant();
			}
		}.bind( this ), 25 );
	},

	resetVariant: function(){
		var variantbase = window.localStorage.getItem( window.relearn.absBaseUri+'/customvariantbase' );
		if( variantbase && confirm( 'You have made changes to your custom variant. Are you sure you want to reset all changes?' ) ){
			window.localStorage.removeItem( window.relearn.absBaseUri+'/customvariantbase' );
			window.localStorage.removeItem( window.relearn.absBaseUri+'/customvariant' );
			this.removeCustomVariantOption();
			if( this.getVariant() == this.customvariantname ){
				this.changeVariant( variantbase );
			}
		}
	},

	onLoadStylesheet: function(){
		variants.isstylesheetloaded = true;
	},

	switchStylesheet: function( variant, without_check ){
		var link = document.querySelector( '#R-variant-style' );
		if( !link ){
			return;
		}
		var old_path = link.getAttribute( 'href' );
		var new_path = this.generateVariantPath( variant, old_path );
		this.isstylesheetloaded = false;

		// Chrome needs a new element to trigger the load callback again
		var new_link = document.createElement( 'link' );
		new_link.id = 'R-variant-style';
		new_link.rel = 'stylesheet';
		new_link.onload = this.onLoadStylesheet;
		new_link.setAttribute( 'href', new_path );
		link.parentNode.replaceChild( new_link, link );
	},

	changeVariant: function( variant ){
		if( variant == this.customvariantname ){
			var variantbase = window.localStorage.getItem( window.relearn.absBaseUri+'/customvariantbase' );
			if( this.variants.indexOf( variantbase ) < 0 ){
				variant = '';
			}
			if( !window.localStorage.getItem( window.relearn.absBaseUri+'/customvariant' ) ){
				variant = '';
			}
			this.setVariant( variant );
			if( !variant ){
				return;
			}
			this.switchStylesheet( variantbase );
			this.loadCustomVariant();
		}
		else{
			if( this.variants.indexOf( variant ) < 0 ){
				variant = this.variants.length ? this.variants[ 0 ] : '';
			}
			this.setVariant( variant );
			if( !variant ){
				return;
			}
			this.switchStylesheet( variant );
			this.markSelectedVariant();
		}
	},

	generator: function( vargenerator ){
		var graphDefinition = this.generateGraph();
		var graphs = document.querySelectorAll( vargenerator );
		graphs.forEach( function( e ){ e.innerHTML = graphDefinition; });

		var interval_id = setInterval( function(){
			if( document.querySelectorAll( vargenerator + ' .mermaid > svg' ).length ){
				clearInterval( interval_id );
				this.styleGraph();
			}
		}.bind( this ), 25 );
	},

	download: function(data, mimetype, filename){
		var blob = new Blob([data], { type: mimetype });
		var url = window.URL.createObjectURL(blob);
		var a = document.createElement('a');
		a.setAttribute('href', url);
		a.setAttribute('download', filename);
		a.click();
	},

	getStylesheet: function(){
		var style = this.generateStylesheet();
		if( !style ){
			alert( 'There is nothing to be generated as auto mode variants will be generated by Hugo' );
			return;
		}
		this.download( style, 'text/css', 'theme-' + this.customvariantname + '.css' );
	},

	adjustCSSRules: function(selector, props, sheets){
		// get stylesheet(s)
		if (!sheets) sheets = [...document.styleSheets];
		else if (sheets.sup){    // sheets is a string
			let absoluteURL = new URL(sheets, document.baseURI).href;
			sheets = [...document.styleSheets].filter(i => i.href == absoluteURL);
		}
		else sheets = [sheets];  // sheets is a stylesheet

		// CSS (& HTML) reduce spaces in selector to one.
		selector = selector.replace(/\s+/g, ' ');
		const findRule = s => [...s.cssRules].reverse().find(i => i.selectorText == selector)
		let rule = sheets.map(findRule).filter(i=>i).pop()

		const propsArr = props.sup
			? props.split(/\s*;\s*/).map(i => i.split(/\s*:\s*/)) // from string
			: Object.entries(props);                              // from Object

		if (rule) for (let [prop, val] of propsArr){
			// rule.style[prop] = val; is against the spec, and does not support !important.
			rule.style.setProperty(prop, ...val.split(/ *!(?=important)/));
		}
		else {
			sheet = sheets.pop();
			if (!props.sup) props = propsArr.reduce((str, [k, v]) => `${str}; ${k}: ${v}`, '');
			sheet.insertRule(`${selector} { ${props} }`, sheet.cssRules.length);
		}
	},

	normalizeColor: function( c ){
		if( !c || !c.trim ){
			return c;
		}
		c = c.trim();
		c = c.replace( /\s*\(\s*/g, "( " );
		c = c.replace( /\s*\)\s*/g, " )" );
		c = c.replace( /\s*,\s*/g, ", " );
		c = c.replace( /0*\./g, "." );
		c = c.replace( / +/g, " " );
		return c;
	},

	getColorValue: function( c ){
		return this.normalizeColor( getComputedStyle( document.documentElement ).getPropertyValue( '--INTERNAL-'+c ) );
	},

	getColorProperty: function( c, read_style ){
		var e = this.findColor( c );
		var p = this.normalizeColor( read_style.getPropertyValue( '--'+c ) ).replace( '--INTERNAL-', '--' );
		return p;
	},

	findLoadedStylesheet: function( id ){
		for( var n = 0; n < document.styleSheets.length; ++n ){
			if( document.styleSheets[n].ownerNode.id == id ){
				var s = document.styleSheets[n];
				if( s.rules && s.rules.length ){
					for( var m = 0; m < s.rules.length; ++m ){
						if( s.rules[m].selectorText == ':root' ){
							return s.rules[m].style;
						}
						if( s.rules[m].cssRules && s.rules[m].cssRules.length ){
							for( var o = 0; o < s.rules[m].cssRules.length; ++o ){
								if( s.rules[m].cssRules[o].selectorText == ':root' ){
									return s.rules[m].cssRules[o].style;
								}
							}
						}
					}
				}
				break;
			}
		}
		return null;
	},

	changeColor: function( c, without_prompt ){
		var with_prompt = !(without_prompt || false);

		var read_style = this.findLoadedStylesheet( 'R-custom-variant-style' );
		var write_style = this.findLoadedStylesheet( 'R-variant-style' );
		if( !read_style ){
			read_style = write_style;
		}
		if( !read_style ){
			if( with_prompt ){
				alert( 'An auto mode variant can not be changed. Please select its light/dark variant directly to make changes' );
			}
			return;
		}

		var e = this.findColor( c );
		var v = this.getColorProperty( c, read_style );
		var n = '';
		if( !with_prompt ){
			n = v;
		}
		else{
			var t = c + '\n\n' + e.tooltip + '\n';
			if( e.fallback ){
				t += '\nInherits value "' + this.getColorValue(e.fallback) + '" from ' + e.fallback + ' if not set\n';
			}
			else if( e.default ){
				t += '\nDefaults to value "' + this.normalizeColor(e.default) + '" if not set\n';
			}
			n = prompt( t, v );
			if( n === null ){
				// user canceld operation
				return;
			}
		}

		if( n ){
			// value set to specific value
			n = this.normalizeColor( n ).replace( '--INTERNAL-', '--' ).replace( '--', '--INTERNAL-' );
			if( !with_prompt || n != v ){
				write_style.setProperty( '--'+c, n );
			}
		}
		else{
			// value emptied, so delete it
			write_style.removeProperty( '--'+c );
		}

		if( with_prompt ){
			this.saveCustomVariant();
		}
	},

	findColor: function( name ){
		var f = this.variantvariables.find( function( x ){
			return x.name == name;
		});
		return f;
	},

	generateColorVariable: function( e, read_style ){
		var v = '';
		var gen = this.getColorProperty( e.name, read_style );
		if( gen ){
			v += '  --' + e.name + ': ' + gen + '; /* ' + e.tooltip + ' */\n';
		}
		return v;
	},

	generateStylesheet: function(){
		var read_style = this.findLoadedStylesheet( 'R-custom-variant-style' );
		var write_style = this.findLoadedStylesheet( 'R-variant-style' );
		if( !read_style ){
			read_style = write_style;
		}
		if( !read_style ){
			return;
		}

		var style =
			'/* ' + this.customvariantname + ' */\n' +
			':root {\n' +
			this.variantvariables.reduce( function( a, e ){ return a + this.generateColorVariable( e, read_style ); }.bind( this ), '' ) +
			'}\n';
		console.log( style );
		return style;
	},

	styleGraphGroup: function( selector, colorvar ){
		this.adjustCSSRules( '#R-body svg '+selector+' > rect', 'color: var(--INTERNAL-'+colorvar+'); fill: var(--INTERNAL-'+colorvar+'); stroke: #80808080;' );
		this.adjustCSSRules( '#R-body svg '+selector+' > .label .nodeLabel', 'color: var(--INTERNAL-'+colorvar+'); fill: var(--INTERNAL-'+colorvar+'); stroke: #80808080;' );
		this.adjustCSSRules( '#R-body svg '+selector+' > .cluster-label .nodeLabel', 'color: var(--INTERNAL-'+colorvar+'); fill: var(--INTERNAL-'+colorvar+'); stroke: #80808080;' );
		this.adjustCSSRules( '#R-body svg '+selector+' .nodeLabel', 'filter: grayscale(1) invert(1) contrast(10000);' );
	},

	styleGraph: function(){
		this.variantvariables.forEach( function( e ){
			this.styleGraphGroup( '.'+e.name, e.name );
		}.bind( this ) );
		this.styleGraphGroup( '#maincontent', 'MAIN-BG-color' );
		this.styleGraphGroup( '#mainheadings', 'MAIN-BG-color' );
		this.styleGraphGroup( '#code', 'CODE-BLOCK-BG-color' );
		this.styleGraphGroup( '#inlinecode', 'CODE-INLINE-BG-color' );
		this.styleGraphGroup( '#blockcode', 'CODE-BLOCK-BG-color' );
		this.styleGraphGroup( '#thirdparty', 'MAIN-BG-color' );
		this.styleGraphGroup( '#coloredboxes', 'BOX-BG-color' );
		this.styleGraphGroup( '#menu', 'MENU-SECTIONS-BG-color' );
		this.styleGraphGroup( '#menuheader', 'MENU-HEADER-BG-color' );
		this.styleGraphGroup( '#menusections', 'MENU-SECTIONS-ACTIVE-BG-color' );
	},

	generateGraphGroupedEdge: function( e ){
		var edge = '';
		if( e.fallback && e.group == this.findColor( e.fallback ).group ){
			edge += e.fallback+':::'+e.fallback+' --> '+e.name+':::'+e.name;
		}
		else{
			edge += e.name+':::'+e.name;
		}
		return edge;
	},

	generateGraphVarGroupedEdge: function( e ){
		var edge = '';
		if( e.fallback && e.group != this.findColor( e.fallback ).group ){
			edge += '  ' + e.fallback+':::'+e.fallback+' --> '+e.name+':::'+e.name + '\n';
		}
		return edge;
	},

	generateGraph: function(){
		var g_groups = {};
		var g_handler = '';

		this.variantvariables.forEach( function( e ){
			var group = e.group || ' ';
			g_groups[ group ] = ( g_groups[ group ] || [] ).concat( e );
			g_handler += '  click '+e.name+' variants.changeColor\n';
		});

		var graph =
			'flowchart LR\n' +
			'  subgraph menu["menu"]\n' +
			'    direction TB\n' +
			'    subgraph menuheader["header"]\n' +
			'      direction LR\n' +
					g_groups[ 'header' ].reduce( function( a, e ){ return a + '      ' + this.generateGraphGroupedEdge( e ) + '\n'; }.bind( this ), '' ) +
			'    end\n' +
			'    subgraph menusections["sections"]\n' +
			'      direction LR\n' +
					g_groups[ 'sections' ].reduce( function( a, e ){ return a + '      ' + this.generateGraphGroupedEdge( e ) + '\n'; }.bind( this ), '' ) +
			'    end\n' +
			'  end\n' +
			'  subgraph maincontent["content"]\n' +
			'    direction TB\n' +
					g_groups[ 'content' ].reduce( function( a, e ){ return a + '    ' + this.generateGraphGroupedEdge( e ) + '\n'; }.bind( this ), '' ) +
			'    subgraph mainheadings["headings"]\n' +
			'      direction LR\n' +
					g_groups[ 'headings' ].reduce( function( a, e ){ return a + '      ' + this.generateGraphGroupedEdge( e ) + '\n'; }.bind( this ), '' ) +
			'    end\n' +
			'    subgraph code["code"]\n' +
			'      direction TB\n' +
					g_groups[ 'code' ].reduce( function( a, e ){ return a + '    ' + this.generateGraphGroupedEdge( e ) + '\n'; }.bind( this ), '' ) +
			'      subgraph inlinecode["inline code"]\n' +
			'        direction LR\n' +
						g_groups[ 'inline code' ].reduce( function( a, e ){ return a + '      ' + this.generateGraphGroupedEdge( e ) + '\n'; }.bind( this ), '' ) +
			'      end\n' +
			'      subgraph blockcode["code blocks"]\n' +
			'        direction LR\n' +
						g_groups[ 'code blocks' ].reduce( function( a, e ){ return a + '      ' + this.generateGraphGroupedEdge( e ) + '\n'; }.bind( this ), '' ) +
			'      end\n' +
			'    end\n' +
			'    subgraph thirdparty["3rd party"]\n' +
			'      direction LR\n' +
					g_groups[ '3rd party' ].reduce( function( a, e ){ return a + '      ' + this.generateGraphGroupedEdge( e ) + '\n'; }.bind( this ), '' ) +
			'    end\n' +
			'    subgraph coloredboxes["colored boxes"]\n' +
			'      direction LR\n' +
					g_groups[ 'colored boxes' ].reduce( function( a, e ){ return a + '      ' + this.generateGraphGroupedEdge( e ) + '\n'; }.bind( this ), '' ) +
			'    end\n' +
			'  end\n' +
				this.variantvariables.reduce( function( a, e ){ return a + this.generateGraphVarGroupedEdge( e ); }.bind( this ), '' ) +
			g_handler;

		console.log( graph );
		return graph;
	},

	variantvariables: [
		{ name: 'PRIMARY-color',                         group: 'content',       fallback: 'MENU-HEADER-BG-color',        tooltip: 'brand primary color', },
		{ name: 'SECONDARY-color',                       group: 'content',       fallback: 'MAIN-LINK-color',             tooltip: 'brand secondary color', },
		{ name: 'ACCENT-color',                          group: 'content',        default: '#ffff00',                     tooltip: 'brand accent color, used for search highlights', },

		{ name: 'MAIN-TOPBAR-BORDER-color',              group: 'content',        default: 'transparent',                 tooltip: 'border color between topbar and content', },
		{ name: 'MAIN-LINK-color',                       group: 'content',       fallback: 'SECONDARY-color',             tooltip: 'link color of content', },
		{ name: 'MAIN-LINK-HOVER-color',                 group: 'content',       fallback: 'MAIN-LINK-color',             tooltip: 'hoverd link color of content', },
		{ name: 'MAIN-BG-color',                         group: 'content',        default: '#ffffff',                     tooltip: 'background color of content', },
		{ name: 'TAG-BG-color',                          group: 'content',       fallback: 'PRIMARY-color',               tooltip: 'tag color', },

		{ name: 'MAIN-TEXT-color',                       group: 'content',        default: '#101010',                     tooltip: 'text color of content and h1 titles', },

		{ name: 'MAIN-TITLES-TEXT-color',                group: 'headings',      fallback: 'MAIN-TEXT-color',             tooltip: 'text color of h2-h6 titles and transparent box titles', },
		{ name: 'MAIN-TITLES-H1-color',                  group: 'headings',      fallback: 'MAIN-TEXT-color',             tooltip: 'text color of h1 titles', },
		{ name: 'MAIN-TITLES-H2-color',                  group: 'headings',      fallback: 'MAIN-TITLES-TEXT-color',      tooltip: 'text color of h2-h6 titles', },
		{ name: 'MAIN-TITLES-H3-color',                  group: 'headings',      fallback: 'MAIN-TITLES-H2-color',        tooltip: 'text color of h3-h6 titles', },
		{ name: 'MAIN-TITLES-H4-color',                  group: 'headings',      fallback: 'MAIN-TITLES-H3-color',        tooltip: 'text color of h4-h6 titles', },
		{ name: 'MAIN-TITLES-H5-color',                  group: 'headings',      fallback: 'MAIN-TITLES-H4-color',        tooltip: 'text color of h5-h6 titles', },
		{ name: 'MAIN-TITLES-H6-color',                  group: 'headings',      fallback: 'MAIN-TITLES-H5-color',        tooltip: 'text color of h6 titles', },

		{ name: 'MAIN-font',                             group: 'content',        default: '"Work Sans", "Helvetica", "Tahoma", "Geneva", "Arial", sans-serif', tooltip: 'text font of content and h1 titles', },

		{ name: 'MAIN-TITLES-TEXT-font',                 group: 'headings',      fallback: 'MAIN-font',                   tooltip: 'text font of h2-h6 titles and transparent box titles', },
		{ name: 'MAIN-TITLES-H1-font',                   group: 'headings',      fallback: 'MAIN-font',                   tooltip: 'text font of h1 titles', },
		{ name: 'MAIN-TITLES-H2-font',                   group: 'headings',      fallback: 'MAIN-TITLES-TEXT-font',       tooltip: 'text font of h2-h6 titles', },
		{ name: 'MAIN-TITLES-H3-font',                   group: 'headings',      fallback: 'MAIN-TITLES-H2-font',         tooltip: 'text font of h3-h6 titles', },
		{ name: 'MAIN-TITLES-H4-font',                   group: 'headings',      fallback: 'MAIN-TITLES-H3-font',         tooltip: 'text font of h4-h6 titles', },
		{ name: 'MAIN-TITLES-H5-font',                   group: 'headings',      fallback: 'MAIN-TITLES-H4-font',         tooltip: 'text font of h5-h6 titles', },
		{ name: 'MAIN-TITLES-H6-font',                   group: 'headings',      fallback: 'MAIN-TITLES-H5-font',         tooltip: 'text font of h6 titles', },

		{ name: 'CODE-theme',                            group: 'code',           default: 'relearn-light',               tooltip: 'name of the chroma stylesheet file', },
		{ name: 'CODE-font',                             group: 'code',           default: '"Consolas", menlo, monospace', tooltip: 'text font of code', },
		{ name: 'CODE-BLOCK-color',                      group: 'code blocks',    default: '#000000',                     tooltip: 'fallback text color of block code; should be adjusted to your selected chroma style', },
		{ name: 'CODE-BLOCK-BG-color',                   group: 'code blocks',    default: '#f8f8f8',                     tooltip: 'fallback background color of block code; should be adjusted to your selected chroma style', },
		{ name: 'CODE-BLOCK-BORDER-color',               group: 'code blocks',   fallback: 'CODE-BLOCK-BG-color',         tooltip: 'border color of block code', },
		{ name: 'CODE-INLINE-color',                     group: 'inline code',    default: '#5e5e5e',                     tooltip: 'text color of inline code', },
		{ name: 'CODE-INLINE-BG-color',                  group: 'inline code',    default: '#fffae9',                     tooltip: 'background color of inline code', },
		{ name: 'CODE-INLINE-BORDER-color',              group: 'inline code',    default: '#fbf0cb',                     tooltip: 'border color of inline code', },

		{ name: 'BROWSER-theme',                         group: '3rd party',      default: 'light',                       tooltip: 'name of the theme for browser scrollbars of the main section', },
		{ name: 'MERMAID-theme',                         group: '3rd party',      default: 'default',                     tooltip: 'name of the default Mermaid theme for this variant, can be overridden in hugo.toml', },
		{ name: 'OPENAPI-theme',                         group: '3rd party',      default: 'light',                       tooltip: 'name of the default OpenAPI theme for this variant, can be overridden in hugo.toml', },
		{ name: 'OPENAPI-CODE-theme',                    group: '3rd party',      default: 'obsidian',                    tooltip: 'name of the default OpenAPI code theme for this variant, can be overridden in hugo.toml', },

		{ name: 'MENU-BORDER-color',                     group: 'header',         default: 'transparent',                 tooltip: 'border color between menu and content', },
		{ name: 'MENU-TOPBAR-BORDER-color',              group: 'header',        fallback: 'MENU-HEADER-BG-color',        tooltip: 'border color of vertical line between menu and topbar', },
		{ name: 'MENU-TOPBAR-SEPARATOR-color',           group: 'header',         default: 'transparent',                 tooltip: 'separator color of vertical line between menu and topbar', },
		{ name: 'MENU-HEADER-BG-color',                  group: 'header',        fallback: 'PRIMARY-color',               tooltip: 'background color of menu header', },
		{ name: 'MENU-HEADER-BORDER-color',              group: 'header',        fallback: 'MENU-HEADER-BG-color',        tooltip: 'border color between menu header and menu', },
		{ name: 'MENU-HEADER-SEPARATOR-color',           group: 'header',        fallback: 'MENU-HEADER-BORDER-color',    tooltip: 'separator color between menu header and menu', },
		{ name: 'MENU-HOME-LINK-color',                  group: 'header',         default: '#323232',                     tooltip: 'home button color if configured', },
		{ name: 'MENU-HOME-LINK-HOVER-color',            group: 'header',         default: '#808080',                     tooltip: 'hoverd home button color if configured', },
		{ name: 'MENU-SEARCH-color',                     group: 'header',         default: '#e0e0e0',                     tooltip: 'text and icon color of search box', },
		{ name: 'MENU-SEARCH-BG-color',                  group: 'header',         default: '#323232',                     tooltip: 'background color of search box', },
		{ name: 'MENU-SEARCH-BORDER-color',              group: 'header',        fallback: 'MENU-SEARCH-BG-color',        tooltip: 'border color of search box', },

		{ name: 'MENU-SECTIONS-BG-color',                group: 'sections',       default: '#282828',                     tooltip: 'background of the menu; this is NOT just a color value but can be a complete CSS background definition including gradients, etc.', },
		{ name: 'MENU-SECTIONS-ACTIVE-BG-color',         group: 'sections',       default: 'rgba( 0, 0, 0, .166 )',       tooltip: 'background color of the active menu section', },
		{ name: 'MENU-SECTIONS-LINK-color',              group: 'sections',       default: '#bababa',                     tooltip: 'link color of menu topics', },
		{ name: 'MENU-SECTIONS-LINK-HOVER-color',        group: 'sections',      fallback: 'MENU-SECTIONS-LINK-color',    tooltip: 'hoverd link color of menu topics', },
		{ name: 'MENU-SECTION-ACTIVE-CATEGORY-color',    group: 'sections',       default: '#444444',                     tooltip: 'text color of the displayed menu topic', },
		{ name: 'MENU-SECTION-ACTIVE-CATEGORY-BG-color', group: 'sections',      fallback: 'MAIN-BG-color',               tooltip: 'background color of the displayed menu topic', },
		{ name: 'MENU-SECTION-ACTIVE-CATEGORY-BORDER-color', group: 'sections',   default: 'transparent',                 tooltip: 'border color between the displayed menu topic and the content', },
		{ name: 'MENU-SECTION-SEPARATOR-color',          group: 'sections',       default: '#606060',                     tooltip: 'separator color between menu sections and menu footer', },
		{ name: 'MENU-VISITED-color',                    group: 'sections',      fallback: 'SECONDARY-color',             tooltip: 'icon color of visited menu topics if configured', },

		{ name: 'BOX-CAPTION-color',                     group: 'colored boxes',  default: 'rgba( 255, 255, 255, 1 )',    tooltip: 'text color of colored box titles', },
		{ name: 'BOX-BG-color',                          group: 'colored boxes',  default: 'rgba( 255, 255, 255, .833 )', tooltip: 'background color of colored boxes', },
		{ name: 'BOX-TEXT-color',                        group: 'colored boxes', fallback: 'MAIN-TEXT-color',             tooltip: 'text color of colored box content', },

		{ name: 'BOX-BLUE-color',                        group: 'colored boxes',  default: 'rgba( 48, 117, 229, 1 )',     tooltip: 'background color of blue boxes', },
		{ name: 'BOX-INFO-color',                        group: 'colored boxes', fallback: 'BOX-BLUE-color',              tooltip: 'background color of info boxes', },
		{ name: 'BOX-BLUE-TEXT-color',                   group: 'colored boxes', fallback: 'BOX-TEXT-color',              tooltip: 'text color of blue boxes', },
		{ name: 'BOX-INFO-TEXT-color',                   group: 'colored boxes', fallback: 'BOX-BLUE-TEXT-color',         tooltip: 'text color of info boxes', },

		{ name: 'BOX-GREEN-color',                       group: 'colored boxes',  default: 'rgba( 42, 178, 24, 1 )',      tooltip: 'background color of green boxes', },
		{ name: 'BOX-TIP-color',                         group: 'colored boxes', fallback: 'BOX-GREEN-color',             tooltip: 'background color of tip boxes', },
		{ name: 'BOX-GREEN-TEXT-color',                  group: 'colored boxes', fallback: 'BOX-TEXT-color',              tooltip: 'text color of green boxes', },
		{ name: 'BOX-TIP-TEXT-color',                    group: 'colored boxes', fallback: 'BOX-GREEN-TEXT-color',        tooltip: 'text color of tip boxes', },

		{ name: 'BOX-GREY-color',                        group: 'colored boxes',  default: 'rgba( 128, 128, 128, 1 )',    tooltip: 'background color of grey boxes', },
		{ name: 'BOX-NEUTRAL-color',                     group: 'colored boxes', fallback: 'BOX-GREY-color',              tooltip: 'background color of neutral boxes', },
		{ name: 'BOX-GREY-TEXT-color',                   group: 'colored boxes', fallback: 'BOX-TEXT-color',              tooltip: 'text color of grey boxes', },
		{ name: 'BOX-NEUTRAL-TEXT-color',                group: 'colored boxes', fallback: 'BOX-GREY-TEXT-color',         tooltip: 'text color of neutral boxes', },

		{ name: 'BOX-ORANGE-color',                      group: 'colored boxes',  default: 'rgba( 237, 153, 9, 1 )',      tooltip: 'background color of orange boxes', },
		{ name: 'BOX-NOTE-color',                        group: 'colored boxes', fallback: 'BOX-ORANGE-color',            tooltip: 'background color of note boxes', },
		{ name: 'BOX-ORANGE-TEXT-color',                 group: 'colored boxes', fallback: 'BOX-TEXT-color',              tooltip: 'text color of orange boxes', },
		{ name: 'BOX-NOTE-TEXT-color',                   group: 'colored boxes', fallback: 'BOX-ORANGE-TEXT-color',       tooltip: 'text color of note boxes', },

		{ name: 'BOX-RED-color',                         group: 'colored boxes',  default: 'rgba( 224, 62, 62, 1 )',      tooltip: 'background color of red boxes', },
		{ name: 'BOX-WARNING-color',                     group: 'colored boxes', fallback: 'BOX-RED-color',               tooltip: 'background color of warning boxes', },
		{ name: 'BOX-RED-TEXT-color',                    group: 'colored boxes', fallback: 'BOX-TEXT-color',              tooltip: 'text color of red boxes', },
		{ name: 'BOX-WARNING-TEXT-color',                group: 'colored boxes', fallback: 'BOX-RED-TEXT-color',          tooltip: 'text color of warning boxes', },
	],
};
