window.relearn = window.relearn || {};

window.relearn.runInitialSearch = function(){
    if( window.relearn.isSearchInit && window.relearn.isLunrInit ){
        var input = document.querySelector('#R-search-by-detail');
        if( !input ){
            return;
        }
        var value = input.value;
        searchDetail( value );
    }
}

var lunrIndex, pagesIndex;

function initLunrIndex( index ){
    pagesIndex = index;
    // Set up Lunr by declaring the fields we use
    // Also provide their boost level for the ranking
    lunrIndex = lunr(function() {
        this.use(lunr.multiLanguage.apply(null, contentLangs));
        this.ref('index');
        this.field('title', {
            boost: 15
        });
        this.field('tags', {
            boost: 10
        });
        this.field('content', {
            boost: 5
        });

        this.pipeline.remove(lunr.stemmer);
        this.searchPipeline.remove(lunr.stemmer);

        // Feed Lunr with each file and let LUnr actually index them
        pagesIndex.forEach(function(page, idx) {
            page.index = idx;
            this.add(page);
        }, this);
    });

    window.relearn.isLunrInit = true;
    window.relearn.runInitialSearch();
}

function triggerSearch(){
    var input = document.querySelector('#R-search-by-detail');
    if( !input ){
        return;
    }
    var value = input.value;
    searchDetail( value );

    // add a new entry to the history after the user
    // changed the term; this does not reload the page
    // but will add to the history and update the address bar URL
    var url = new URL( window.location );
    var oldValue = url.searchParams.get( 'search-by' );
    if( value != oldValue ){
        var state = window.history.state || {};
        state = Object.assign( {}, ( typeof state === 'object' ) ? state : {} );
        url.searchParams.set( 'search-by', value );
        state.search = url.toString();
        // with normal pages, this is handled by the 'pagehide' event, but this
        // doesn't fire in case of pushState, so we have to do the same thing
        // here, too
        state.contentScrollTop = +elc.scrollTop;
        window.history.pushState( state, '', url );
    }
}

window.addEventListener( 'popstate', function ( event ){
    // restart search if browsed through history
    if( event.state ){
        var state = window.history.state || {};
        state = Object.assign( {}, ( typeof state === 'object' ) ? state : {} );
        if( state.search ) {
            var url = new URL( state.search );
            if( url.searchParams.has('search-by') ){
                var search = url.searchParams.get( 'search-by' );

				// we have to insert the old search term into the inputs
                var inputs = document.querySelectorAll( 'input.search-by' );
                inputs.forEach( function( e ){
                    e.value = search;
                    var event = document.createEvent( 'Event' );
                    event.initEvent( 'input', false, false );
                    e.dispatchEvent( event );
                });

				// recreate the last search results and eventually
				// restore the previous scrolling position
                searchDetail( search );
            }
        }
    }
});

var input = document.querySelector('#R-search-by-detail');
if( input ){
    input.addEventListener( 'keydown', function(event) {
        // if we are pressing ESC in the searchdetail our focus will
        // be stolen by the other event handlers, so we have to refocus
        // here after a short while
        if (event.key == "Escape") {
            setTimeout( function(){ input.focus(); }, 0 );
        }
    });
}

function initLunrJs() {
    // new way to load our search index
    if( window.index_js_url ){
        var js = document.createElement("script");
        js.src = index_js_url;
        js.setAttribute("async", "");
        js.onload = function(){
            initLunrIndex(relearn_search_index);
        };
        js.onerror = function(e){
            console.error('Error getting Hugo index file');
        };
        document.head.appendChild(js);
    }
}

/**
 * Trigger a search in Lunr and transform the result
 *
 * @param  {String} term
 * @return {Array}  results
 */
function search(term) {
    // Find the item in our index corresponding to the Lunr one to have more info
    // Remove Lunr special search characters: https://lunrjs.com/guides/searching.html
    term = term.replace( /[*:^~+-]/g, ' ' );
    var searchTerm = lunr.tokenizer( term ).reduce( function(a,token){return a.concat(searchPatterns(token.str))}, []).join(' ');
    return !searchTerm || !lunrIndex ? [] : lunrIndex.search(searchTerm).map(function(result) {
        return { index: result.ref, matches: Object.keys(result.matchData.metadata) }
    });
}

function searchPatterns(word) {
    // for short words high amounts of typos doesn't make sense
    // for long words we allow less typos because this largly increases search time
    var typos = [
        { len:  -1, typos: 1 },
        { len:  60, typos: 2 },
        { len:  40, typos: 3 },
        { len:  20, typos: 4 },
        { len:  16, typos: 3 },
        { len:  12, typos: 2 },
        { len:   8, typos: 1 },
        { len:   4, typos: 0 },
    ];
    return [
        word + '^100',
        word + '*^10',
        '*' + word + '^10',
        word + '~' + typos.reduce( function( a, c, i ){ return word.length < c.len ? c : a; } ).typos + '^1'
    ];
}


function resolvePlaceholders( s, args ) {
    var args = args || [];
    // use replace to iterate over the string
    // select the match and check if the related argument is present
    // if yes, replace the match with the argument
    return s.replace(/{([0-9]+)}/g, function (match, index) {
        // check if the argument is present
        return typeof args[index] == 'undefined' ? match : args[index];
    });
};

function searchDetail( value ) {
    var results = document.querySelector('#R-searchresults');
    var hint = document.querySelector('.searchhint');
    hint.innerText = '';
    results.textContent = '';
    var a = search( value );
    if( a.length ){
        hint.innerText = resolvePlaceholders( window.T_N_results_found, [ value, a.length ] );
        a.forEach( function(item){
            var page = pagesIndex[item.index];
            var numContextWords = 10;
            var contextPattern = '(?:\\S+ +){0,' + numContextWords + '}\\S*\\b(?:' +
                item.matches.map( function(match){return match.replace(/\W/g, '\\$&')} ).join('|') +
                ')\\b\\S*(?: +\\S+){0,' + numContextWords + '}';
            var context = page.content.match(new RegExp(contextPattern, 'i'));
            var divsuggestion = document.createElement('a');
            divsuggestion.className = 'autocomplete-suggestion';
            divsuggestion.setAttribute('data-term', value);
            divsuggestion.setAttribute('data-title', page.title);
            divsuggestion.setAttribute('href', window.relearn.relBaseUri + page.uri);
            divsuggestion.setAttribute('data-context', context);
            var divtitle = document.createElement('div');
            divtitle.className = 'title';
            divtitle.innerText = '» ' + page.title;
            divsuggestion.appendChild( divtitle );
            var divbreadcrumb = document.createElement('div');
            divbreadcrumb.className = 'breadcrumbs';
            divbreadcrumb.innerText = (page.breadcrumb || '');
            divsuggestion.appendChild( divbreadcrumb );
            if( context ){
                var divcontext = document.createElement('div');
                divcontext.className = 'context';
                divcontext.innerText = (context || '');
                divsuggestion.appendChild( divcontext );
            }
            results.appendChild( divsuggestion );
        });
        window.relearn.markSearch();
    }
    else if( value.length ) {
        hint.innerText = resolvePlaceholders( window.T_No_results_found, [ value ] );
    }
    input.focus();
    setTimeout( adjustContentWidth, 0 );

	// if we are initiating search because of a browser history
	// operation, we have to restore the scrolling postion the
	// user previously has used; if this search isn't initiated
	// by a browser history operation, it simply does nothing
    var state = window.history.state || {};
    state = Object.assign( {}, ( typeof state === 'object' ) ? state : {} );
    if( state.hasOwnProperty( 'contentScrollTop' ) ){
        window.setTimeout( function(){
            elc.scrollTop = +state.contentScrollTop;
        }, 10 );
        return;
    }
}

initLunrJs();

function startSearch(){
    var input = document.querySelector('#R-search-by-detail');
    if( input ){
        var state = window.history.state || {};
        state = Object.assign( {}, ( typeof state === 'object' ) ? state : {} );
        state.search = window.location.toString();
        window.history.replaceState( state, '', window.location );
    }

    var searchList = new autoComplete({
        /* selector for the search box element */
        selectorToInsert: 'search:has(.searchbox)',
        selector: '#R-search-by',
        /* source is the callback to perform the search */
        source: function(term, response) {
            response(search(term));
        },
        /* renderItem displays individual search results */
        renderItem: function(item, term) {
            var page = pagesIndex[item.index];
            var numContextWords = 2;
            var contextPattern = '(?:\\S+ +){0,' + numContextWords + '}\\S*\\b(?:' +
                item.matches.map( function(match){return match.replace(/\W/g, '\\$&')} ).join('|') +
                ')\\b\\S*(?: +\\S+){0,' + numContextWords + '}';
            var context = page.content.match(new RegExp(contextPattern, 'i'));
            var divsuggestion = document.createElement('div');
            divsuggestion.className = 'autocomplete-suggestion';
            divsuggestion.setAttribute('data-term', term);
            divsuggestion.setAttribute('data-title', page.title);
            divsuggestion.setAttribute('data-uri', window.relearn.relBaseUri + page.uri);
            divsuggestion.setAttribute('data-context', context);
            var divtitle = document.createElement('div');
            divtitle.className = 'title';
            divtitle.innerText = '» ' + page.title;
            divsuggestion.appendChild( divtitle );
            if( context ){
                var divcontext = document.createElement('div');
                divcontext.className = 'context';
                divcontext.innerText = (context || '');
                divsuggestion.appendChild( divcontext );
            }
            return divsuggestion.outerHTML;
        },
        /* onSelect callback fires when a search suggestion is chosen */
        onSelect: function(e, term, item) {
            location.href = item.getAttribute('data-uri');
            e.preventDefault();
        }
    });
};

ready( startSearch );
